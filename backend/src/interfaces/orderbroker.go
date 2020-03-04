package interfaces

// The Order Broker is a component that's useless in this implementation,
// but I wanted to illustrate what types of components could be
// linked into the process of handling a request. In this case, one reason for an order broker
// is rate limiting on the application layer.

type OrderBroker struct {
	BaseOrderHandler
}

func CreateOrderBroker() *OrderBroker {
	return &OrderBroker{}
}
