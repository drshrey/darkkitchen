package interfaces

import "fmt"

type Dispatcher struct {
	BaseOrderHandler
	darkKitchen *DarkKitchen
}

func CreateDispatcher(darkKitchen *DarkKitchen) *Dispatcher {
	return &Dispatcher{
		darkKitchen: darkKitchen,
	}
}

func (d *Dispatcher) HandleOrder(order Order) error {
	if d.nextOrderHandler != nil {
		err := d.nextOrderHandler.HandleOrder(order)
		if err != nil {
			return err
		} else {
			d.dispatchDriver(order)
		}
	} else {
		return fmt.Errorf("nextOrderHandler is nil")
	}

	return nil
}

// emulating driver response
func (d *Dispatcher) dispatchDriver(order Order) {
	// in a production system, we would create a request for a driver
	// from one of our partner systems e.g. UberEATS, DoorDash that would then find a driver and
	// send us a "driver found" response. For simplicity, we'll directly create the driver that should receive the order
	driver := CreateDriver(d.darkKitchen)
	driver.ReceiveOrderRequest(order)
}
