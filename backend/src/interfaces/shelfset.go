package interfaces

import (
	"fmt"
	"sort"
)

var OVERFLOW_SHELF_SIZE = 20

// ShelfSet implements CarrierFacility. It contains the functionality
// and representation of the shelves within our DarkKitchen.
type ShelfSet struct {
	shelves map[string][]Order
	BaseOrderHandler
	countNoSpace int
	countDecay   int
	// this channel receives UUIDs that match
	// orders within the ShelfSet
	orderDeathNotifications chan Order
	shutdownMonitor         chan bool
	darkKitchen             *DarkKitchen
}

// Used for referencing an order in a particular shelf
type ShelfOrder struct {
	ShelfLabel string
	ShelfIndex int
	Order
}

func CreateShelfSet(darkKitchen *DarkKitchen) *ShelfSet {
	hotOrders := make([]Order, 15)
	coldOrders := make([]Order, 15)
	frozenOrders := make([]Order, 15)
	overflowOrders := make([]Order, OVERFLOW_SHELF_SIZE)
	orderDeathNotifications := make(chan Order, ORDER_DEATH_NOTIFICATIONS_SIZE)
	shutdownMonitor := make(chan bool, 1)

	s := &ShelfSet{
		shelves: map[string][]Order{
			HOT_TEMPERATURE_LABEL:    hotOrders,
			COLD_TEMPERATURE_LABEL:   coldOrders,
			FROZEN_TEMPERATURE_LABEL: frozenOrders,
			OVERFLOW_LABEL:           overflowOrders,
		},
		orderDeathNotifications: orderDeathNotifications,
		shutdownMonitor:         shutdownMonitor,
		darkKitchen:             darkKitchen,
	}

	return s
}

// GetState packages the shelf state into a parsable output
// like { "hot": [{ orderObj, ... }], }
func (s *ShelfSet) GetState() interface{} {
	shelfState := map[string]interface{}{}

	type ViewableOrder struct {
		ID               string  `json:"id"`
		Name             string  `json:"name"`
		NormalizedHealth float32 `json:"normalizedHealth"`
		Temperature      string  `json:"temp"`
		PickedUp         bool    `json:"pickedUp"`
	}

	for label := range s.shelves {
		shelfState[label] = []ViewableOrder{}
		for _, order := range s.shelves[label] {
			if order != nil {
				shelfState[label] = append(shelfState[label].([]ViewableOrder), ViewableOrder{
					ID:               order.GetID(),
					Name:             order.GetName(),
					NormalizedHealth: order.GetHealth() / order.GetShelfLife(),
					Temperature:      order.GetTemperature(),
					PickedUp:         order.GetPickedUp(),
				})
			}
		}
	}

	shelfState[SHELFSET_WASTED_ORDERS_DECAY_LABEL] = s.countDecay
	shelfState[SHELFSET_WASTED_ORDERS_NOSPACE_LABEL] = s.countNoSpace

	return shelfState
}

func (s *ShelfSet) HandleOrder(order Order) error {
	// add order to shelf and start a goroutine for that Order
	// which runs the decay process
	err := s.AddOrderToShelf(order)
	if err != nil {
		s.countNoSpace++
		return err
	}

	return nil
}

// Run as goroutine for shelf-set to monitor
// when orders go to waste. This way, we can fill in
// the empty space thereafter with an order from the overflow
// shelf, if possible
func (s *ShelfSet) MonitorOrderDecayNotifications() {
	for {
		select {
		case wastedOrder := <-s.orderDeathNotifications:
			// find wastedOrder in our shelves
			orderTemp := wastedOrder.GetTemperature()

			orderFound := false
			if s.shelves[orderTemp] != nil {
				for idx, order := range s.shelves[orderTemp] {
					if order != nil && order.GetID() == wastedOrder.GetID() {
						// orderFound = true
						// remove order off of shelf
						s.removeOrder(orderTemp, idx)
						s.countDecay++
						orderFound = true

						// Check if we can add an Order from the overflow shelf
						overflowOrder, err := s.GetShortestLivingOrderFromOverflowShelf(orderTemp)
						if err == nil {
							// assign order to empty space and remove from overflow shelf
							s.addOrder(overflowOrder.Order, orderTemp, idx)
							s.removeOrder(OVERFLOW_LABEL, overflowOrder.ShelfIndex)
						}
					}
				}
			}

			// check overflow shelf if order not found on temperature shelf
			if !orderFound {
				for idx, order := range s.shelves[OVERFLOW_LABEL] {
					if order != nil && order.GetID() == wastedOrder.GetID() {
						// remove order off of shelf
						orderFound = true
						s.removeOrder(OVERFLOW_LABEL, idx)
						s.countDecay++
					}
				}
			}

		case _ = <-s.shutdownMonitor:
			break
		}
	}
}

func (s *ShelfSet) Start() {
	go s.MonitorOrderDecayNotifications()
}

func (s *ShelfSet) Shutdown() {
	s.shutdownMonitor <- true
}

func (s *ShelfSet) removeOrder(label string, idx int) {
	s.shelves[label][idx] = nil
	s.darkKitchen.CarrierFacilityHasBeenUpdated()
}

func (s *ShelfSet) addOrder(order Order, label string, idx int) {
	s.shelves[label][idx] = order

	if label == OVERFLOW_LABEL {
		order.SetCurrentDecayRate(float32(OVERFLOW_PREMIUM) * order.GetOriginalDecayRate())
	} else {
		order.SetCurrentDecayRate(order.GetOriginalDecayRate())
	}

	// send notification to DarkKitchen there's been an update to the shelf
	s.darkKitchen.CarrierFacilityHasBeenUpdated()
}

// Finds the first available empty space from the particular shelf of the ShelfSet
func (s *ShelfSet) GetEmptySpaceFromShelf(shelfLabel string) (*int, error) {
	outIdx := 0
	for idx, shelfOrder := range s.shelves[shelfLabel] {
		if shelfOrder == nil {
			outIdx = idx
			return &outIdx, nil
		}
	}

	return nil, fmt.Errorf(NoSpaceLeftOnShelfErr)
}

// Finds the Order with the highest health in the particular shelf
func (s *ShelfSet) GetHighestHealthOrderFromShelf(shelfLabel string) *ShelfOrder {
	var highestHealthOrder *ShelfOrder
	for idx, shelfOrder := range s.shelves[shelfLabel] {
		// if shelf is full, let's find the order with the highest health to insert into
		// the overflow shelf if that's possible
		if highestHealthOrder == nil || (shelfOrder != nil && shelfOrder.GetHealth() < highestHealthOrder.GetHealth()) {
			highestHealthOrder = &ShelfOrder{
				ShelfIndex: idx,
				ShelfLabel: shelfLabel,
				Order:      shelfOrder,
			}
		}
	}

	return highestHealthOrder
}

// AddOrderToShelf looks for the appropriate shelf
// to append the order to and adds it to that shelf
// if possible
func (s *ShelfSet) AddOrderToShelf(order Order) error {
	emptySpaceFound := false
	var highestHealthOrder *ShelfOrder

	switch order.GetTemperature() {
	// The code structure for all cases is the same with the difference being
	// the label that is used and the requested order to be added to the shelf set.
	// Since we are mutating the state of the shelves, we want to show the
	// explicit nature of the addition rather than hiding
	// it within a function and reducing the number of code replication
	case HOT_TEMPERATURE_LABEL:
		emptySpaceIdx, err := s.GetEmptySpaceFromShelf(HOT_TEMPERATURE_LABEL)
		if err != nil {
			// find highest health order to place in overflow shelf
			highestHealthOrder = s.GetHighestHealthOrderFromShelf(HOT_TEMPERATURE_LABEL)
		} else {
			emptySpaceFound = true
			s.addOrder(order, HOT_TEMPERATURE_LABEL, *emptySpaceIdx)

			s.darkKitchen.WG.Add(1)
			go order.Decay(s.orderDeathNotifications, getShelfDecayValue)
		}
	case COLD_TEMPERATURE_LABEL:
		emptySpaceIdx, err := s.GetEmptySpaceFromShelf(COLD_TEMPERATURE_LABEL)
		if err != nil {
			highestHealthOrder = s.GetHighestHealthOrderFromShelf(COLD_TEMPERATURE_LABEL)
		} else {
			emptySpaceFound = true
			s.addOrder(order, COLD_TEMPERATURE_LABEL, *emptySpaceIdx)

			s.darkKitchen.WG.Add(1)
			go order.Decay(s.orderDeathNotifications, getShelfDecayValue)
		}
	case FROZEN_TEMPERATURE_LABEL:
		emptySpaceIdx, err := s.GetEmptySpaceFromShelf(FROZEN_TEMPERATURE_LABEL)
		if err != nil {
			highestHealthOrder = s.GetHighestHealthOrderFromShelf(FROZEN_TEMPERATURE_LABEL)
		} else {
			emptySpaceFound = true
			s.addOrder(order, FROZEN_TEMPERATURE_LABEL, *emptySpaceIdx)

			s.darkKitchen.WG.Add(1)
			go order.Decay(s.orderDeathNotifications, getShelfDecayValue)
		}
	default:
		return fmt.Errorf(ShelfWithLabelNotFoundErr, order.GetTemperature())
	}

	// if all temperature shelves are empty
	if !emptySpaceFound {
		// check for an empty space in the overflow shelf
		for idx, shelfOrder := range s.shelves[OVERFLOW_LABEL] {
			if shelfOrder == nil {
				emptySpaceFound = true

				// if the highestHealthOrder has a health gt the input order,
				// move that to overflow. otherwise, insert the input order into overflow
				if highestHealthOrder.Order.GetHealth() > order.GetHealth() {
					s.removeOrder(highestHealthOrder.ShelfLabel, highestHealthOrder.ShelfIndex)
					// add highesthealth order into overflow shelf
					s.addOrder(highestHealthOrder.Order, OVERFLOW_LABEL, idx)

					// add input order into temperature shelf now that there is space
					s.addOrder(order, highestHealthOrder.ShelfLabel, highestHealthOrder.ShelfIndex)
				} else {
					// add input order into overflow
					s.addOrder(order, OVERFLOW_LABEL, idx)
				}

				s.darkKitchen.WG.Add(1)
				go order.Decay(s.orderDeathNotifications, getShelfDecayValue)

				break
			}
		}

		if !emptySpaceFound {
			return fmt.Errorf(NoSpaceLeftErr)
		}
	}

	return nil
}

// GetShortestLivingOrderFromOverflowShelf returns the
// order with least life left from the specified
// temperature classification, if exists
func (s *ShelfSet) GetShortestLivingOrderFromOverflowShelf(temp string) (*ShelfOrder, error) {
	sortedOrders := []ShelfOrder{}
	for idx, order := range s.shelves[OVERFLOW_LABEL] {
		if order != nil && order.GetTemperature() == temp {
			sortedOrders = append(sortedOrders, ShelfOrder{
				ShelfLabel: OVERFLOW_LABEL,
				ShelfIndex: idx,
				Order:      order,
			})
		}
	}

	if len(sortedOrders) > 0 {
		// sort Orders by health
		sort.Sort(byHealth(sortedOrders))
		return &sortedOrders[len(sortedOrders)-1], nil
	} else {
		return nil, fmt.Errorf("No orders with temp %s found", temp)
	}
}

// GiveOrder finds the order with the input orderID
// and returns that back, if exists
func (s *ShelfSet) GiveOrder(orderID string) (Order, error) {
	var foundOrder Order

	for shelfLabel := range s.shelves {
		for shelfIndex, order := range s.shelves[shelfLabel] {
			if order != nil && order.GetID() == orderID {
				// empty out shelf space
				foundOrder = order
				s.removeOrder(shelfLabel, shelfIndex)

				// Check if we can add an Order from the overflow shelf
				if shelfLabel != OVERFLOW_LABEL {
					overflowOrder, err := s.GetShortestLivingOrderFromOverflowShelf(shelfLabel)
					if err == nil {
						// assign order to empty space and
						// remove from overflow shelf
						s.addOrder(overflowOrder.Order, shelfLabel, shelfIndex)
						s.removeOrder(OVERFLOW_LABEL, overflowOrder.ShelfIndex)
					}
				}

				break
			}
		}
	}

	if foundOrder != nil {
		foundOrder.SetPickedUp(true)
		return foundOrder, nil
	} else {
		return nil, fmt.Errorf("No order found for id: %s", orderID)
	}
}

func getShelfDecayValue(shelfLife float32, orderAge float32, decayRate float32) float32 {
	return (shelfLife - orderAge) - (decayRate * orderAge)
}

// Implement Sort interface to use byHealth as a custom sorting function - https://golang.org/pkg/sort/#Interface
type byHealth []ShelfOrder

func (s byHealth) Len() int {
	return len(s)
}

// ascending where Orders with most life
// are in the beginning
func (s byHealth) Less(i, j int) bool {
	if s[i].GetHealth() > s[j].GetHealth() {
		return true
	} else {
		return false
	}
}

func (s byHealth) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
