package interfaces

import (
	"sync"
)

// implements ProcessingCenter
type DarkKitchen struct {
	OrderBroker     *OrderBroker
	Kitchen         *Kitchen
	Dispatcher      *Dispatcher
	CarrierFacility CarrierFacility
	// used for managing driver threads and shelfset decay process thread
	// this is so the program does not exit until all goroutines have completed execution
	WG           *sync.WaitGroup
	WastedOrders int
	// Used for sending notifications to the websocket handler
	// to return the most updated state of the shelves to the client
	// where the notification is the name of the CK component
	// currently only supports "carrierfacility"
	UpdatedStateNotifications chan string

	// simulation config stores the configuration information
	// for simulating any parts of the order request process
	// that requires simulating e.g. the driver min and max delay times
	simulationConfig *SimulationConfig
}

// CreateDarkKitchen creates a base DarkKitchen instance. If it is being used
// for simulation purposes, allow the user to pass in the simulation config as a valid parameter
func CreateDarkKitchen(simulationConfig *SimulationConfig) *DarkKitchen {

	darkKitchen := &DarkKitchen{}

	orderBroker := CreateOrderBroker()
	kitchen := CreateKitchen()
	dispatcher := CreateDispatcher(darkKitchen)
	carrierFacility := CreateShelfSet(darkKitchen)

	updatedStateNotifications := make(chan string, 10000)

	darkKitchen.OrderBroker = orderBroker
	darkKitchen.Kitchen = kitchen
	darkKitchen.Dispatcher = dispatcher
	darkKitchen.CarrierFacility = carrierFacility
	darkKitchen.WG = &sync.WaitGroup{}
	darkKitchen.WastedOrders = 0
	darkKitchen.UpdatedStateNotifications = updatedStateNotifications
	darkKitchen.simulationConfig = simulationConfig

	// start any running processes for the carrier facility e.g.
	// process for handling decayed Orders
	carrierFacility.Start()

	darkKitchen.OrderBroker.SetNextOrderHandler(darkKitchen.Kitchen)
	darkKitchen.Kitchen.SetNextOrderHandler(darkKitchen.Dispatcher)
	darkKitchen.Dispatcher.SetNextOrderHandler(darkKitchen.CarrierFacility)

	return darkKitchen
}

func (ck *DarkKitchen) ReceiveOrder(order Order) error {
	err := ck.OrderBroker.HandleOrder(order)
	if err != nil {
		return err
	}

	return nil
}

func (ck *DarkKitchen) CarrierFacilityHasBeenUpdated() {
	ck.UpdatedStateNotifications <- CARRIER_FACILITY_LABEL
}
