package interfaces_test

import (
	"testing"
	"time"

	"github.com/drshrey/darkkitchen/backend/src/interfaces"
)

// Testing the order creation flow
func TestReceiveOrder_Success(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)
	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 100, interfaces.HOT_TEMPERATURE_LABEL, ck)
	err := ck.ReceiveOrder(&newOrder)
	if err != nil {
		t.Error(err)
	}

	if newOrder.GetName() != "order-name" {
		t.Error("Name does not match")
	}

	if newOrder.GetTemperature() != interfaces.HOT_TEMPERATURE_LABEL {
		t.Error("Temperature does not match")
	}

	// check if order has been picked up
	// and confirm that it does not die
	for {
		if newOrder.GetPickedUp() {
			break
		}

		if newOrder.GetHealth() <= 0 {
			t.Error("Order died")
		}
	}

	ck.WG.Wait()
}

func TestReceiveOrder_Failure_GetWastedOrdersThatCannotBeAddedToShelves(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(100, 100, 10*time.Millisecond)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	wastedOrdersCount := 0
	for i := 0; i < 40; i++ {
		ck.WG.Add(1)
		go func(idx int) {
			defer ck.WG.Done()
			newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 100, interfaces.HOT_TEMPERATURE_LABEL, ck)
			err := ck.ReceiveOrder(&newOrder)
			if err != nil {
				wastedOrdersCount++
				return
			}

			// check if order has been picked up
			// and confirm that it does not die
			for {
				if newOrder.GetPickedUp() {
					break
				}

				if newOrder.GetHealth() <= 0 {
					break
				}
			}

			return
		}(i)
	}

	ck.WG.Wait()

	// max is 35 since temp shelf(15) + overflow_shelf(20) = 35
	if wastedOrdersCount != 5 {
		t.Error("wasted order count is not exactly 5")
	}
}

func TestReceiveOrder_Success_AllTemperatureOrders(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(20, 20, 10)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	for _, label := range []string{interfaces.HOT_TEMPERATURE_LABEL, interfaces.COLD_TEMPERATURE_LABEL, interfaces.FROZEN_TEMPERATURE_LABEL} {
		for i := 0; i < 15; i++ {
			ck.WG.Add(1)
			go func(idx int, tempLabel string) {
				defer ck.WG.Done()
				newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 100, tempLabel, ck)
				err := ck.ReceiveOrder(&newOrder)
				if err != nil {
					t.Error(err)
					return
				}

				// check if order has been picked up
				// and confirm that it does not die
				for {
					if newOrder.GetPickedUp() {
						break
					}

					if newOrder.GetHealth() <= 0 {
						t.Error("Order died")
						return
					}
				}

				return
			}(i, label)
		}
	}

	ck.WG.Wait()
}

func TestReceiveOrder_Success_WastedOrder(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(100, 100, 10)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	newOrder := interfaces.CreateFoodOrder("order-name", 10, 1, interfaces.HOT_TEMPERATURE_LABEL, ck)
	ck.ReceiveOrder(&newOrder)

	if newOrder.GetName() != "order-name" {
		t.Error("Name does not match")
	}

	// check if order has been picked up
	// and confirm that it does not die
	for {
		if newOrder.GetPickedUp() {
			t.Error("Order was picked up")
			break
		}

		if newOrder.GetHealth() <= 0 {
			break
		}
	}

	ck.WG.Wait()
}

func TestCreateOrder_Failure_InvalidTemperature(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(20, 20, 10*time.Millisecond)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 10, "INVALID-TEMP", ck)
	err := ck.ReceiveOrder(&newOrder)
	if err == nil {
		t.Error("Order was accepted")
	}

	ck.WG.Wait()
}

// Test Driver related functionality
func TestDriverReceiveOrderAtPickupPoint_Failure_OrderMismatch(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	testCf := TestCarrierFacility{}
	ck.CarrierFacility = testCf

	driver := interfaces.CreateDriver(ck)
	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 10, "INVALID-TEMP", ck)

	err := driver.ReceiveOrderRequest(&newOrder)
	if err == nil {
		t.Error("Order was accepted")
	}
}

func TestDriverReceiveOrderRequest_Failure_NilCarrierFacility(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)
	ck.CarrierFacility = nil

	driver := interfaces.CreateDriver(ck)
	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 10, "INVALID-TEMP", ck)

	err := driver.ReceiveOrderRequest(&newOrder)
	if err == nil {
		t.Error("Order was accepted")
	}
}

// Test BaseOrderHandler related functionality
func TestBaseOrderHandlerHandleOrder_Failure_NilNextOrderHandler(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	boh := interfaces.BaseOrderHandler{}
	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 10, "INVALID-TEMP", ck)
	err := boh.HandleOrder(&newOrder)
	if err == nil {
		t.Error("HandleOrder should error out since nil nextOrderHandler")
	}
}

// Test Dispatcher related functionality
func TestDispatcherHandleOrder_Failure_NilNextOrderHandler(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	dispatcher := interfaces.CreateDispatcher(ck)
	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 10, "INVALID-TEMP", ck)
	err := dispatcher.HandleOrder(&newOrder)
	if err == nil {
		t.Error("HandleOrder should error out since nil nextOrderHandler")
	}
}

// Test ShelfSet related functionality
func TestShelfSetGetState_Success(t *testing.T) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)
	shelfSet := interfaces.CreateShelfSet(ck)

	// add order to shelf
	newOrder := interfaces.CreateFoodOrder("order-name", 0.1, 10, "INVALID-TEMP", ck)
	shelfSet.AddOrderToShelf(&newOrder)

	state := shelfSet.GetState()

	typeAssertedState, ok := state.(map[string]interface{})
	if !ok {
		t.Error("state not expected type")
	}

	if typeAssertedState[interfaces.OVERFLOW_LABEL] == nil {
		t.Error("overflow shelf does not exist")
	}

	if typeAssertedState[interfaces.HOT_TEMPERATURE_LABEL] == nil {
		t.Error("hot shelf does not exist")
	}

	if typeAssertedState[interfaces.COLD_TEMPERATURE_LABEL] == nil {
		t.Error("cold shelf does not exist")
	}

	if typeAssertedState[interfaces.FROZEN_TEMPERATURE_LABEL] == nil {
		t.Error("frozen shelf does not exist")
	}
}

// Test Interfaces
type TestCarrierFacility struct {
	interfaces.CarrierFacility
}

func (t TestCarrierFacility) GiveOrder(orderID string) (interfaces.Order, error) {
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	ck := interfaces.CreateDarkKitchen(simulationConfig)

	// create invalid order and return that
	invalidOrder := interfaces.CreateFoodOrder("somename", 0.1, 10, interfaces.HOT_TEMPERATURE_LABEL, ck)
	return &invalidOrder, nil
}
