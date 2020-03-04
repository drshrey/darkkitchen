## The Dark Kitchen Simulation Environment

This document outlines the design of the Dark Kitchen Simulation Environment. The Dark Kitchen Simulation Environment is a system that, given orders in the specified Order format via HTTP requests, processes orders in real-time and displays the state of the Dark Kitchen i.e. the normalized value of each order split up by their respective shelf. We will also show the total waste that has been produced by the system in order to highlight the estimated costs associated with processing the following orders. This is a simulation because although the design of the system is flexible enough to be used for processing real-world orders and dispatching requests to actual drivers, we are simulating the behavior of those agents and their environment in this particular implementation.

### Running the Code

#### Pre-requisites

##### Running it

1. Docker. Install it here: https://docs.docker.com/install/.
2. Docker Compose. Once you've installed Docker, install docker compose here: https://docs.docker.com/compose/install/.

3. `cd` to the root directory of this repository
4. `docker-compose build && docker-compose up`

5. Go to http://localhost:3000
6. Go to the client folder and run `docker-compose build && docker-compose up`
7. Go back to http://localhost:3000 to see the state of the Dark Kitchen shelves
8. To re-test, click "Clean Simulation Environment" and follow steps 1-3 again.

##### Cleaning up

1. Once you're finished, you can `CTRL+c` in the same window you ran step 2 and confirm all containers are removed by doing `docker-compose rm -f`
2. `docker ps` to validate there are no more containers on your Docker host

##### Running tests

1. `docker-compose -f docker-compose.test.yml build && docker-compose -f docker-compose.test.yml up` in root directory

### Software Architecture 

#### Design Problems

1. *Managing the flow of processing an order in an isolated manner*
2. *Abstracting the order entity such that it can be used for other types of orders in the future*
3. *Proper threading for Order decay processes and having a decoupled design between the order and shelf*

### System Design

We take an object-oriented approach here because it serves as a much more direct approach to representating the real-world scenario we wish to model versus a function-oriented design.

- [Interaction Diagram](https://www.dropbox.com/s/iypydarfsp114so/interaction_diagram%20%282%29.png?dl=0)

#### Concepts

- Interfaces (in `backend/src/interfaces.go`)
1. `Order` provides an interface for the general-purpose properties and methods of what an order should contain.
- Implementations: `FoodOrder`
2. `CarrierFacility` is the interface for systems that house and manage completed orders before they are given off to the `Couriers`
- Implementations: `ShelfSet`
3. `Courier` is the interface for agents that would pick up `Orders` from the `CarrierFacility`.
- Implementations: `Driver`
4. `OrderHandler` is for an entity that processes an order and needs to pass it off to the next entity in the process chain. This is useful as we want to decouple the processes that do
something with an order from each other to allow flexible process chains.
- Implementations: `Kitchen`, `Dispatcher`, `OrderBroker`

**Constraints:**

- Only when a shelf has free space, you can move an Order back from the overflow shelf to that particular shelf.
- A new order must be placed on its correct shelf when it is initially handled by the `ShelfSet`. Only if its correct shelf is full can we put it on the overflow shelf.

**Process:**

The situations in which we consider moving Orders **to** the Overflow shelf are:

- When an Order is considered to be placed on its initial shelf and its correct temperature shelf is full. Here, we find the Order with the highest health including ones on the temperature shelf and the new Order. We then send that to the overflow shelf if the overflow shelf is not also full.

The situations in which we consider moving Orders **from** the Overflow shelf are:

- When an Order decays to 0 and is from a temperature shelf, we check for an Order in the overflow shelf to see if we can fill the new empty space in the temperature shelf. We find the Order on the overflow shelf that has the least health for that particular temperature. This way, if there are multiple orders on the overflow shelf of the same temperature, we get the one that would be most affected if its decay rate was to go back to normal.
- When an Order from a temperature shelf is given to the driver that requests their Order, we go through the same process of finding the "shortest life left" Order to replace the Order that's gone away.
- When an Order is requested to be added to the `ShelfSet`, we see if there is empty space available in its respective Temperature shelf. If there is no space, we get the Order with the highest health of that temperature (including the Order that's being requested), and send it to the overflow shelf, if possible. If the Order added to the overflow shelf is an existing order from the temperature shelf, we then fill the now empty space with the currently requested Order.

### Technologies Used

*Backend*

- Golang, Gorilla for websockets

*Frontend*

- ReactJS

*Infra*

- Docker

### Index

The **Order** format:

```json
{
    "name" : "Cheese Pizza",
    "temp" : "hot",
    "shelfLife" : 300,
    "decayRate" : 0.45
}
```

**Directory Structure**

```
.
├── backend
│   ├── Dockerfile
│   ├── Dockerfile.test
│   ├── Gopkg.lock
│   ├── Gopkg.toml
│   └── src
│       ├── interfaces
│       │   ├── darkkitchen.go
│       │   ├── cover.out
│       │   ├── dispatcher.go
│       │   ├── driver.go
│       │   ├── interfaces.go
│       │   ├── interfaces_test.go
│       │   ├── kitchen.go
│       │   ├── order.go
│       │   ├── orderbroker.go
│       │   ├── shelfset.go
│       │   └── variables.go
│       ├── main.go
│       └── testdata
│           └── orders.json
├── client
│   ├── Dockerfile
│   ├── Gopkg.lock
│   ├── Gopkg.toml
│   └── src
│       └── main.go
├── docker-compose.test.yml
├── docker-compose.yml
└── frontend
    ├── Dockerfile
    ├── README.md
    ├── package.json
    ├── public
    │   ├── favicon.ico
    │   ├── index.html
    │   └── manifest.json
    ├── run.sh
    ├── src
    │   ├── App.css
    │   ├── App.js
    │   ├── index.css
    │   ├── index.js
    └── yarn.lock

9 directories, 38 files
```

