package interfaces

// Kitchen implements OrderHandler by subclassing BaseOrderHandler. Although
// there isn't anything custom in this implementation, we may want to implement methods
// like CookOrder and other Kitchen related functions that don't exist at the moment.
type Kitchen struct {
	BaseOrderHandler
}

func CreateKitchen() *Kitchen {
	return &Kitchen{}
}
