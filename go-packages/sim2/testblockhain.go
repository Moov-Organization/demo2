package sim2

import (
	"reflect"
	"fmt"
	"sync"
	"strings"
)

type TestChain struct {
	recvChans []chan string
	sendChans []chan string
	currentRides []Ride
	requestedRides []Ride
	RecvServer chan Ride
	mutex *sync.Mutex
}

type Ride struct {
	from string
	to string
}

func NewTestChain() *TestChain {
	tc := new(TestChain)
	tc.RecvServer = make(chan Ride)
	tc.mutex = &sync.Mutex{}
	return tc
}

func (tc *TestChain) StartTestChain() {
	go tc.receiveRideRequestsThread()
	go tc.blockchainInteractorsThread()
}


func (tc *TestChain) RegisterBlockchainInteractor()(testchainApi *TestChainAPI){
	testchainApi = new(TestChainAPI)
	testchainApi.recvChan = make(chan string)
	testchainApi.sendChan = make(chan string)
	ride := new(Ride)
	tc.recvChans = append(tc.recvChans, testchainApi.sendChan)
	tc.sendChans = append(tc.sendChans, testchainApi.recvChan)
	tc.currentRides = append(tc.currentRides, *ride)
	return
}


func (tc *TestChain)receiveRideRequestsThread() {
	for {
		ride := <-tc.RecvServer
		tc.mutex.Lock()
		tc.requestedRides = append(tc.requestedRides, ride)
		tc.mutex.Unlock()
	}
}

func (tc *TestChain)blockchainInteractorsThread() {
	for {
		cases := make([]reflect.SelectCase, len(tc.recvChans))
		for i, ch := range tc.recvChans {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		}

		remaining := len(cases)
		for remaining > 0 {
			chosen, value, ok := reflect.Select(cases)
			if !ok {
				// The chosen channel has been closed, so zero out the channel to disable the case
				cases[chosen].Chan = reflect.ValueOf(nil)
				remaining -= 1
				continue
			}

			switch value.String() {
			case "GetRides":
				if len(tc.requestedRides) > 0 {
					tc.sendChans[chosen] <- "true"
				} else {
					tc.sendChans[chosen] <- "false"
				}
			case "AcceptRide":
				if len(tc.requestedRides) > 0 {
					tc.currentRides[chosen] = tc.requestedRides[0]
					tc.mutex.Lock()
					tc.requestedRides = append(tc.requestedRides[:0], tc.requestedRides[1:]...)
					tc.mutex.Unlock()
					tc.sendChans[chosen] <- "true"
				} else {
					tc.sendChans[chosen] <- "false"
				}
			case "GetLocations":
				tc.sendChans[chosen] <- fmt.Sprintf("%s %s", tc.currentRides[chosen].from, tc.currentRides[chosen].to)
			}

		}
	}
}


type TestChainAPI struct {
	recvChan chan string
	sendChan chan string
}

func (testChainApi *TestChainAPI) GetRideAddressIfAvailable() (available bool, address string) {
	testChainApi.sendChan <- "GetRides"
	return "true" == <-testChainApi.recvChan, ""
}

func (testChainApi *TestChainAPI) AcceptRequest(address string) (status bool) {
	testChainApi.sendChan <- "AcceptRide"
	return "true" == <-testChainApi.recvChan
}

func (testChainApi *TestChainAPI) GetLocations(address string) (from string, to string) {
	testChainApi.sendChan <- "GetLocations"
	locations := strings.Split(<-testChainApi.recvChan, " ")
	return locations[0], locations[1]
}
