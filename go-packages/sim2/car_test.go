package sim2

import (
	"testing"
)

func TestNewCar(t *testing.T) {
	var mockEth MockEthAPI
	var mockWorld MockWorld
	var syncChan chan bool
	var recvChan *chan CarInfo
	mockWorld.getRandomEdgeStruct.returnEdge =
		Edge{Start:&Vertex{Pos:Coords{0,0}},
			End:&Vertex{Pos:Coords{1,1}}}
	car := NewCar(0, &mockWorld, &mockEth, syncChan, recvChan)
	if car.id != 0 {
		t.Errorf("Car ID does not equal %d \n", 0)
	}
	if car.path.pathState != DrivingAtRandom {
		t.Errorf("Car path state not intialized to random \n")
	}
}
