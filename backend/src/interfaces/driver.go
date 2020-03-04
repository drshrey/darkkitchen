package interfaces

import (
	"fmt"
	"math/rand"
	"time"
)

// Driver implements Courier
type Driver struct {
	etaToCarrierFacility int
	messages             chan string
	hasPickedUpOrder     bool
	OrderRequest         Order
	darkKitchen          *DarkKitchen
}

func CreateDriver(darkKitchen *DarkKitchen) Driver {
	messagesChan := make(chan string, 1000)
	return Driver{
		messages:    messagesChan,
		darkKitchen: darkKitchen,
	}
}

// ReceiveOrder receives the order
// from the carrier facility that is housing
// the order the driver wishes to pick up
func (d *Driver) ReceiveOrder() error {
	if d.darkKitchen.CarrierFacility == nil {
		return fmt.Errorf(NilCarrierFacilityErr)
	}

	order, err := d.darkKitchen.CarrierFacility.GiveOrder(d.OrderRequest.GetID())
	if err != nil {
		return err
	}

	if order != nil && order.GetID() == d.OrderRequest.GetID() {
		d.hasPickedUpOrder = true
		d.DeliverOrder()
	} else {
		return fmt.Errorf("Given order is not the same as requested")
	}

	return nil
}

func (d *Driver) ReceiveOrderRequest(request Order) error {
	d.OrderRequest = request

	d.darkKitchen.WG.Add(1)
	go d.startDriverJourney()

	msg := ""
	for {
		select {
		case driverResolution := <-d.messages:
			msg = driverResolution
			break
		}

		if d.etaToCarrierFacility <= 0 {
			break
		}
	}

	if msg != "received" {
		return fmt.Errorf("%s", msg)
	}

	return nil
}

// DeliverOrder has no actual logic implemented in this case
// but necessary for implementing the Courier interface
func (d *Driver) DeliverOrder() {
	return
}

// simulate driver journey by sleeping
// for the time that it would take
// to travel to carrier facility
func (d *Driver) startDriverJourney() {
	defer d.darkKitchen.WG.Done()

	if d.darkKitchen.simulationConfig == nil {
		d.messages <- NoSimulationConfigErr
		return
	}

	d.etaToCarrierFacility = d.darkKitchen.simulationConfig.DriverMinDelay + rand.Intn(d.darkKitchen.simulationConfig.DriverMaxDelay)
	for {
		time.Sleep(d.darkKitchen.simulationConfig.SleepTime)
		d.etaToCarrierFacility--
		if d.etaToCarrierFacility == 0 {
			break
		}
	}

	err := d.ReceiveOrder()
	if err != nil {
		d.messages <- err.Error()
	} else {
		d.messages <- DRIVER_RECEIVED_MSG
	}
}
