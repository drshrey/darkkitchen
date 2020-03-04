package interfaces

import "fmt"

type Order interface {
	GetID() string
	GetName() string
	GetTemperature() string
	GetCurrentDecayRate() float32
	SetCurrentDecayRate(float32)
	GetOriginalDecayRate() float32
	GetShelfLife() float32
	GetOrderAge() float32
	SetOrderAge(float32)
	GetHealth() float32
	GetPickedUp() bool
	SetPickedUp(bool)
	// pass in decay func w/ (shelfLife, orderAge, decayRate) format
	Decay(chan Order, func(float32, float32, float32) float32)
}

type OrderHandler interface {
	SetNextOrderHandler(OrderHandler)
	HandleOrder(Order) error
}

type Courier interface {
	ReceiveOrderRequest(Order, CarrierFacility) error
	ReceiveOrderAtPickupPoint() error
	DeliverOrder() error
}

/*
	CarrierFacility is for components
	that manage Orders before given to
	the respective agents for delivering the Order
	to the customer. This is different from regular OrderHandlers
	because rather than processing and making modifications to the Order,
	they are simply managing their state.
*/
type CarrierFacility interface {
	OrderHandler
	GiveOrder(string) (Order, error)
	GetState() interface{}
	Start()
	Shutdown()
}

// base class for OrderHandler
type BaseOrderHandler struct {
	nextOrderHandler OrderHandler
}

func (b *BaseOrderHandler) SetNextOrderHandler(handler OrderHandler) {
	b.nextOrderHandler = handler
}

func (b *BaseOrderHandler) HandleOrder(order Order) error {
	if b.nextOrderHandler != nil {
		err := b.nextOrderHandler.HandleOrder(order)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("nextOrderHandler is nil")
	}

	return nil
}
