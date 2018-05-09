package sim2

import (
	"testing"
	"time"
)
var mockEth MockEthAPI
var mockWorld MockWorld
var syncChan chan bool
var recvChan *chan CarInfo
var car *Car

func reset() {
	mockEth = *new(MockEthAPI)
	mockWorld = *new(MockWorld)
	car = new(Car)
	car.ethApi = &mockEth
	car.world = &mockWorld
	car.path.currentEdge = Edge{ID:0,
		Start:&Vertex{Pos:Coords{0,0}},
		End:&Vertex{Pos:Coords{1,1}}}
	car.path.currentPos = Coords{0,0}
}

func TestNewCar(t *testing.T) {
	mockWorld.getRandomEdgeStruct.returnEdge =
		Edge{ID:0,
			Start:&Vertex{Pos:Coords{0,0}},
			End:&Vertex{Pos:Coords{1,1}}}
	car = NewCar(0, &mockWorld, &mockEth, syncChan, recvChan)
	if car.id != 0 {
		t.Errorf("Car ID does not equal %d \n", 0)
	}
	if car.path.pathState != DrivingAtRandom {
		t.Errorf("Car path state not intialized to random \n")
	}
	if car.requestState != None {
		t.Errorf("Car request state not intialized to none \n")
	}
}

func TestCar_CarLoop(t *testing.T) {
	testRequestState(t)
	testPathState(t)
}

func testRequestState(t *testing.T) {

	// Test for accept request success
	reset()
	mockEth.getAddressStruct.returnAvailable = true
	mockEth.acceptRequestStruct.function = func (string) bool {
		time.Sleep(time.Millisecond * 1)
		return true
	}
	car.requestState = None
	car.drive()
	if car.requestState != Trying {
		t.Errorf("Car request state did not switch to trying \n")
	}
	if mockEth.getAddressStruct.calls != 1 {
		t.Errorf("Did not try to get address in None state \n")
	}
	time.Sleep(time.Millisecond * 2)
	if car.requestState != Success {
		t.Errorf("Car request state did not switch to success \n")
	}

	// Test for accept request failure
	reset()
	mockEth.getAddressStruct.returnAvailable = true
	mockEth.acceptRequestStruct.function = func (string) bool {
		time.Sleep(time.Millisecond * 1)
		return false
	}
	car.requestState = None
	car.drive()
	if car.requestState != Trying {
		t.Errorf("Car request state did not switch to trying \n")
	}
	if mockEth.getAddressStruct.calls != 1 {
		t.Errorf("Did not try to get address in None state \n")
	}
	time.Sleep(time.Millisecond * 2)
	if car.requestState != Fail {
		t.Errorf("Car request state did not switch to success \n")
	}

	// Test for retry after failure
	reset()
	mockEth.getAddressStruct.returnAvailable = true
	mockEth.acceptRequestStruct.function = func (string) bool {
		time.Sleep(time.Millisecond * 1)
		return true
	}
	car.requestState = Fail
	car.drive()
	if car.requestState != Trying {
		t.Errorf("Car request state did not switch to trying \n")
	}
	if mockEth.getAddressStruct.calls != 1 {
		t.Errorf("Did not try to get address in None state \n")
	}
	time.Sleep(time.Millisecond * 2)
	if car.requestState != Success {
		t.Errorf("Car request state did not switch to success \n")
	}

	// Test for procedures after successful accept
	reset()
	car.requestState = Success
	mockEth.getLocationStruct.returnFrom = "0,0"
	mockEth.getLocationStruct.returnTo = "1,1"
	mockWorld.closestEdgeAndCoordStruct.returnLocation = Location{edge:Edge{Start:&Vertex{ID:1}}}
	car.drive()
	if mockEth.getLocationStruct.calls != 1 {
		t.Errorf("Did not query eth api to get locations \n")
	}
	if mockWorld.closestEdgeAndCoordStruct.calls != 2 {
		t.Errorf("Did not query the world twice to find closest location to pick up and drop off points \n")
	}
	if mockWorld.shortestpathStruct.calls != 1 {
		t.Errorf("Did not query the world to find shortest location between current location and pick up location \n")
	}
	if car.path.pathState != ToPickUp {
		t.Errorf("Car path not changed to pick up \n")
	}
	if car.requestState != None {
		t.Errorf("Car request state did not switch to none \n")
	}
}

func testPathState(t *testing.T) {

}
