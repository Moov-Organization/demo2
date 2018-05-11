package sim2

import (
	"testing"
	"time"
	"math"
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
	edge := Edge{ID:0,
		Start:&Vertex{Pos:Coords{0,0}},
		End:&Vertex{Pos:Coords{1,1}}}
	mockWorld.getVertexStruct.returnVertex = *edge.Start
	mockWorld.getVertexStruct.returnVertex.AdjEdges = append(mockWorld.getVertexStruct.returnVertex.AdjEdges, edge)
	mockWorld.getRandomEdgeStruct.returnEdge = edge
	var carID uint = 0
	car = NewCar(carID, &mockWorld, &mockEth, syncChan, recvChan)
	if car.id != carID {
		t.Errorf("Car ID does not equal %d \n", carID)
	}
	if mockWorld.getVertexStruct.paramId != carID {
		t.Errorf("Edge ID does not equal %d \n", carID)
	}
	if mockWorld.shortestpathStruct.calls != 1 {
		t.Errorf("Did not query the world to find shortest location between current location and pick up location \n")
	}
	if car.path.currentState != DrivingAtRandom {
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
	acceptRequestWaitChannel :=  make(chan bool, 1)
	mockEth.acceptRequestStruct.function = func (string) bool {
		return <-acceptRequestWaitChannel
	}
	car.requestState = None
	car.drive()
	if car.requestState != Trying {
		t.Errorf("Car request state did not switch to trying \n")
	}
	if mockEth.getAddressStruct.calls != 1 {
		t.Errorf("Did not try to get address in None state \n")
	}
	acceptRequestWaitChannel <- true
	time.Sleep(time.Millisecond)
	if car.requestState != Success {
		t.Errorf("Car request state did not switch to success \n")
	}

	// Test for accept request failure
	reset()
	mockEth.getAddressStruct.returnAvailable = true
	mockEth.acceptRequestStruct.function = func (string) bool {
		return <-acceptRequestWaitChannel
	}
	car.requestState = None
	car.drive()
	if car.requestState != Trying {
		t.Errorf("Car request state did not switch to trying \n")
	}
	if mockEth.getAddressStruct.calls != 1 {
		t.Errorf("Did not try to get address in None state \n")
	}
	acceptRequestWaitChannel <- false
	time.Sleep(time.Millisecond)
	if car.requestState != Fail {
		t.Errorf("Car request state did not switch to fail \n")
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
	if car.path.currentState != ToPickUp {
		t.Errorf("Car path not changed to pick up \n")
	}
	if car.requestState != None {
		t.Errorf("Car request state did not switch to none \n")
	}
}

func testPathState(t *testing.T) {

	//// Test destination reached after driving at random
	reset()
	mockWorld.getRandomEdgeStruct.returnEdge = car.path.currentEdge
	car.path.currentState = DrivingAtRandom
	car.path.currentPos = car.path.currentEdge.End.Pos
	car.path.routeEdges = []Edge{{ID:1, Start:&Vertex{Pos:Coords{1,1}}, End:&Vertex{Pos:Coords{2,2}}}}
	car.drive()
	if car.path.currentEdge.ID != 1 {
		t.Errorf("Car did not switch edges upon reaching \n")
	}
	if len(car.path.routeEdges) != 0 {
		t.Errorf("Last edge was not popped after reaching destination \n")
	}
	if mockWorld.shortestpathStruct.calls != 1 {
		t.Errorf("Did not query the world to find shortest location between current location and random location \n")
	}
	if car.path.currentState != DrivingAtRandom {
		t.Errorf("Car path state not remained in random \n")
	}

	// Test destination reached after driving to pickup
	reset()
	car.path.dropOff.edge =
		Edge{ID:0,
			Start:&Vertex{Pos:Coords{0,0}},
			End:&Vertex{Pos:Coords{1,1}}}
	car.path.currentState = ToPickUp
	car.path.currentPos = car.path.currentEdge.End.Pos
	car.path.routeEdges = []Edge{{ID:1, Start:&Vertex{Pos:Coords{1,1}}, End:&Vertex{Pos:Coords{2,2}}}}

	car.drive()
	if car.path.currentState != ToDropOff {
		t.Errorf("Car path state not switched to drop off \n")
	}

	// Test destination reached after driving to pickup
	reset()
	mockWorld.getRandomEdgeStruct.returnEdge = car.path.currentEdge
	car.path.currentState = ToDropOff
	car.path.currentPos = car.path.currentEdge.End.Pos
	car.path.routeEdges = []Edge{{ID:1, Start:&Vertex{Pos:Coords{1,1}}, End:&Vertex{Pos:Coords{2,2}}}}
	car.drive()
	if car.path.currentState != DrivingAtRandom {
		t.Errorf("Car path state not switched to driving at random \n")
	}

	// Test if current position is projected to end position by Movement per drive when distance is greater than movement per drive
	reset()
	originalCarPostion := car.path.currentPos
	car.drive()
	distanceMoved := originalCarPostion.Distance(car.path.currentPos)
	if math.Abs(distanceMoved -MovementPerFrame) > 0.1 {
		t.Errorf("Current position was not projected by movement per drive \n")
	}

	// Test if current position is set to end position when distance is less than Movement per drive
	reset()
	car.path.currentEdge.End.Pos = Coords{X:0.1, Y:0.1}
	car.drive()
	if car.path.currentPos != car.path.currentEdge.End.Pos {
		t.Errorf("Current position not set to end position upon getting close enough \n")
	}

}

