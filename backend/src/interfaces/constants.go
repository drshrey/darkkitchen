package interfaces

import "time"

const (
	// in milliseconds
	DEFAULT_SLEEP_TIME                   = 1000 * time.Millisecond
	OVERFLOW_PREMIUM                     = 2
	OVERFLOW_LABEL                       = "overflow"
	HOT_TEMPERATURE_LABEL                = "hot"
	COLD_TEMPERATURE_LABEL               = "cold"
	FROZEN_TEMPERATURE_LABEL             = "frozen"
	CARRIER_FACILITY_LABEL               = "carrierfacility"
	ORDER_DEATH_NOTIFICATIONS_SIZE       = 10000
	DEFAULT_DRIVER_MIN_DELAY             = 2
	DEFAULT_DRIVER_MAX_DELAY             = 8
	SHELFSET_WASTED_ORDERS_DECAY_LABEL   = "wastedOrdersDecay"
	SHELFSET_WASTED_ORDERS_NOSPACE_LABEL = "wastedOrdersNoSpace"
	DRIVER_RECEIVED_MSG                  = "received"
)
