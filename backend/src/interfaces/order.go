package interfaces

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type FoodOrderInput struct {
	Name        string  `json:"name"`
	DecayRate   float32 `json:"decayRate"`
	ShelfLife   float32 `json:"shelfLife"`
	Temperature string  `json:"temp"`
}

// FoodOrder implements Order. In this particular case, we are handling food orders, so
// in the case of custom properties for an order containing food like the decay rate, we want to have a different
// implementation to distinguish from a different order type.
type FoodOrder struct {
	id                string
	name              string
	currentDecayRate  float32
	originalDecayRate float32
	// health is used for future-proofing the effects that other systems
	// may have on the food order e.g. weather conditions. although we only are affected
	// by the decay value function of the shelfset in this case, this may change over time
	// and we want a way to centralize the total effects on the food order
	health      float32
	shelfLife   float32
	temperature string
	orderAge    float32
	pickedUp    bool
	darkKitchen *DarkKitchen
}

func CreateFoodOrder(name string, decayRate float32, shelfLife float32, temperature string, darkKitchen *DarkKitchen) FoodOrder {
	uniqueID := uuid.NewV4()
	return FoodOrder{
		id:                uniqueID.String(),
		name:              name,
		originalDecayRate: decayRate,
		shelfLife:         shelfLife,
		temperature:       temperature,
		orderAge:          0.0,
		health:            shelfLife,
		pickedUp:          false,
		darkKitchen:       darkKitchen,
	}
}

// GetID
func (f *FoodOrder) GetID() string {
	return f.id
}

// GetName
func (f *FoodOrder) GetName() string {
	return f.name
}

// GetOriginalDecayRate
func (f *FoodOrder) GetOriginalDecayRate() float32 {
	return f.originalDecayRate
}

// GetOriginalDecayRate
func (f *FoodOrder) GetCurrentDecayRate() float32 {
	return f.currentDecayRate
}

// GetShelfLife
func (f *FoodOrder) GetShelfLife() float32 {
	return f.shelfLife
}

// GetTemperature
func (f *FoodOrder) GetTemperature() string {
	return f.temperature
}

// GetOrderAge
func (f *FoodOrder) GetOrderAge() float32 {
	return f.orderAge
}

// GetHealth
func (f *FoodOrder) GetHealth() float32 {
	return f.health
}

// GetPickedUp
func (f *FoodOrder) GetPickedUp() bool {
	return f.pickedUp
}

// SetOrderAge
func (f *FoodOrder) SetOrderAge(newOrderAge float32) {
	f.orderAge = newOrderAge
}

// SetCurrentDecayRate
func (f *FoodOrder) SetCurrentDecayRate(newDecayRate float32) {
	f.currentDecayRate = newDecayRate
}

// SetPickedUp shows the state for whether an order has been picked up.
// For instance, we stop the Order's Decay process if an order has been picked up
func (f *FoodOrder) SetPickedUp(pickedUp bool) {
	f.pickedUp = pickedUp
}

// Decay
func (f *FoodOrder) Decay(decayNotifications chan Order, decayValueFn func(shelfLife float32, orderAge float32, decayRate float32) float32) {
	defer f.darkKitchen.WG.Done()

	for {
		time.Sleep(f.darkKitchen.simulationConfig.SleepTime)
		f.SetOrderAge(f.GetOrderAge() + 1)
		// get delta between the current health and decay so we preserve the history of changed decay rates
		// for instance, if the decay rate was 1 but then became 2 once it went to the overflow shelf, we want
		// the health to reflect the total sum of how much it was affected when it was at 1 and how much it was affected when it was at 2
		delta := f.health - decayValueFn(f.GetShelfLife(), f.GetOrderAge(), f.GetCurrentDecayRate())
		f.health -= delta
		if f.health <= 0 {
			// notify the channel that this order has died
			decayNotifications <- f
			break
		}

		if f.GetPickedUp() {
			break
		}
	}

	return
}
