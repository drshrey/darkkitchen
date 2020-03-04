package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/drshrey/darkkitchen/backend/src/interfaces"
	"github.com/gorilla/websocket"
)

func main() {
	// Initialize DarkKitchen with simulation config variables for driver delays
	simulationConfig := interfaces.CreateSimulationConfig(interfaces.DEFAULT_DRIVER_MIN_DELAY, interfaces.DEFAULT_DRIVER_MAX_DELAY, interfaces.DEFAULT_SLEEP_TIME)
	darkKitchen := interfaces.CreateDarkKitchen(simulationConfig)

	// used for handling client order requests
	http.HandleFunc("/orders/new", func(w http.ResponseWriter, r *http.Request) {
		HandleOrderRequest(w, r, darkKitchen)
	})

	// sends the state of the dark kitchen back to the client
	// through a websocket connection
	http.HandleFunc("/ws/darkKitchenState", func(w http.ResponseWriter, r *http.Request) {
		WSDarkKitchenState(w, r, darkKitchen)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func HandleOrderRequest(w http.ResponseWriter, r *http.Request, darkKitchen *interfaces.DarkKitchen) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	requestParams := interfaces.FoodOrderInput{}
	err = json.Unmarshal(body, &requestParams)
	if err != nil {
		logrus.Error(err.Error())
		w.Write([]byte(err.Error()))
		w.WriteHeader(400)
		return
	}

	newOrder := interfaces.CreateFoodOrder(requestParams.Name, requestParams.DecayRate, requestParams.ShelfLife, requestParams.Temperature, darkKitchen)
	darkKitchen.ReceiveOrder(&newOrder)
}

func WSDarkKitchenState(w http.ResponseWriter, r *http.Request, darkKitchen *interfaces.DarkKitchen) {
	// initialize the websocket connection with Gorilla's Upgrader - http://www.gorillatoolkit.org/pkg/websocket#Upgrader.Upgrade
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// for the purposes of this exercise, we accept requests from
			// any client origin
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	defer conn.Close()
	for {
		select {
		case component := <-darkKitchen.UpdatedStateNotifications:
			if component == interfaces.CARRIER_FACILITY_LABEL {
				// get updated shelf state and send it to the client
				cfState := darkKitchen.CarrierFacility.GetState()
				jsonCfState, err := json.Marshal(cfState)
				if err != nil {
					logrus.Error(err.Error())
				} else {
					if err := conn.WriteMessage(websocket.TextMessage, jsonCfState); err != nil {
						return
					}
				}
			}
		}
	}
}

type SimulationRequest struct {
	PoissonRateParameter float32 `json:"poissonRateParam"`
	DriverMinDelay       int     `json:"driverMinDelay"`
	DriverMaxDelay       int     `json:"driverMaxDelay"`
	TimeUnits            int     `json:"timeUnits"`
}
