package interfaces

import "time"

type SimulationConfig struct {
	DriverMinDelay int
	DriverMaxDelay int
	SleepTime      time.Duration
}

// CreateSimulationConfig initializes
// the config object to be used specifically
// for the dark kitchen simulation
func CreateSimulationConfig(driverMinDelay int, driverMaxDelay int, sleepTime time.Duration) *SimulationConfig {
	return &SimulationConfig{
		DriverMaxDelay: driverMinDelay,
		DriverMinDelay: driverMaxDelay,
		SleepTime:      sleepTime,
	}
}
