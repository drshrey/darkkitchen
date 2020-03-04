package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	rng "github.com/leesper/go_rng"
)

type FoodOrderInput struct {
	Name        string  `json:"name"`
	DecayRate   float32 `json:"decayRate"`
	ShelfLife   float32 `json:"shelfLife"`
	Temperature string  `json:"temp"`
}

const DEFAULT_POISSON_RATE_PARAM = 3.25

// this is a simple client implementation using a poisson scale of DEFAULT deliveries per second
// to benchmark our DarkKitchen implementation

func main() {
	// read in JSON file of orders, and call ReceiveOrder for each one
	jsonFile, err := os.Open("/tmp/orders.json")
	if err != nil {
		logrus.Errorf(err.Error())
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var orders []FoodOrderInput
	json.Unmarshal(byteValue, &orders)

	poissonDistribution := rng.NewPoissonGenerator(int64(len(orders)))

	numOrdersInSecond := poissonDistribution.Poisson(DEFAULT_POISSON_RATE_PARAM)
	count := 0

	wg := sync.WaitGroup{}

	for i := 0; i < len(orders); i++ {
		if int64(count) != numOrdersInSecond {
			wg.Add(1)
			go func(orderIdx int) {
				defer wg.Done()
				// make client request
				jsonOrderString, err := json.Marshal(orders[orderIdx])
				if err != nil {
					logrus.Fatal(err)
				}

				logrus.Infof("Processing %s", orders[orderIdx].Name)

				req, err := http.NewRequest("POST", "http://host.docker.internal:8080/orders/new", bytes.NewBuffer(jsonOrderString))
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}

				defer resp.Body.Close()
			}(i)

			count++
		} else {
			numOrdersInSecond = poissonDistribution.Poisson(DEFAULT_POISSON_RATE_PARAM)
			count = 0
			time.Sleep(1 * time.Second)
		}
	}

	wg.Wait()
}
