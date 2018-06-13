package sim2

import (
  "time"
  "strconv"

)

// world - Describes world state: position and velocity of all cars in simulation

type TrafficInfo struct {
	carStates  []CarInfo
	stopLights []StopLightInfo
}

// CarInfo - struct to contain position and velocity information for a simulated car.
type CarInfo struct {
  ID uint
  Pos Coords
  Vel Coords  // with respect to current position, offset for a single frame
  Dir float64  // unit vector with respect to current position
  EdgeId uint
}

type StopLightInfo struct {
	ID uint
	lightstates [NumberOfDirections]LightState
	alarm time.Time
}

type LightState int
const (
	Red     LightState = 0
	Orange  LightState = 1
	Green   LightState = 2
)

// World - struct to contain all relevat world information in simulation.
type World struct {
  graph *Digraph
	trafficInfo TrafficInfo
  fps float64
  numRegisteredCars uint  // Total count of registered cars
  syncChans []chan TrafficInfo  // Index by actor ID for channel to/from that actor
  recvChan chan CarInfo  // Receive from all Cars registered on one channel
  webChan chan Message
}

// NewWorld - Constructor for valid World object.
func NewWorld(fps float64, graph *Digraph) *World {
  w := new(World)
  w.graph = graph
  w.fps = fps
  w.numRegisteredCars = 0

  for _, intersection := range graph.Intersections {
  	if intersection.intersectionType == StopLight {
			w.trafficInfo.stopLights = append(w.trafficInfo.stopLights, StopLightInfo{ID:intersection.id})
		}
	}
  // NOTE recvChan is nil until cars are registered
  // NOTE webChan is nil until registered
  return w
}

// RegisterCar - If the car ID has not been taken, allocate new channels for the car ID and true OK.
//   If the car ID is taken or World is unallocated, return nil channels and false OK value.
func (w *World) RegisterCar() (uint, chan TrafficInfo, *chan CarInfo, bool) {
  // Check for invalid World
  if w == nil {
    return 0, nil, nil, false
  }
  ID := w.numRegisteredCars

  w.numRegisteredCars++
  //fmt.Println("numRegisteredCars:", w.numRegisteredCars)

  // Allocate new channels for registered car
  w.trafficInfo.carStates = append(w.trafficInfo.carStates, CarInfo{ ID:ID }) // TODO: randomize/control car location on startup
  w.syncChans = append(w.syncChans, make(chan TrafficInfo, 1)) // Buffer up to one output
  w.recvChan = make(chan CarInfo, w.numRegisteredCars)  // Overwrite buffered allocation
  return ID, w.syncChans[ID], &w.recvChan, true
}

// TODO: an UnregisterCar func if necessary

// LoopWorld - Begin the world simulation execution loop
func (w *World) LoopWorld() {
	for idx := range w.trafficInfo.stopLights {
		w.trafficInfo.stopLights[idx].lightstates[West] = Green
		w.trafficInfo.stopLights[idx].alarm = time.Now().Add(time.Second * 5)
	}

	itercounter := uint64(0)
  for {
    //fmt.Println("Iteration", itercounter)
    timer := time.NewTimer(time.Duration(1000/w.fps) * time.Millisecond)

		w.updateStopLights()
    // Send out sync flag = true for each registered car
    for ID := range w.trafficInfo.carStates {
      cpyCarInfo := make([]CarInfo, len(w.trafficInfo.carStates))
      copy(cpyCarInfo, w.trafficInfo.carStates)
			cpyStopLightInfo := make([]StopLightInfo, len(w.trafficInfo.stopLights))
			copy(cpyStopLightInfo, w.trafficInfo.stopLights)
      w.syncChans[ID] <- TrafficInfo{cpyCarInfo,cpyStopLightInfo}
    }

    // Car coroutines should now process current world state
    for idx, car := range w.trafficInfo.carStates {
      w.webChan <- Message{
        Type:"Car",
        ID:strconv.Itoa(int(idx)),
        X:strconv.Itoa(int(car.Pos.X)),
        Y:strconv.Itoa(int(car.Pos.Y)),
        Orientation:strconv.Itoa(int(car.Dir)),
      }
    }



    // Wait for all registered cars to report
    for carRecvCt := uint(0); carRecvCt < w.numRegisteredCars ; {
      data := <-w.recvChan
      // TODO: deep copy is safer here
			w.trafficInfo.carStates[data.ID] = data

      carRecvCt++
      //fmt.Println("World got new data on index", data.ID, ":", data)
    }

    itercounter++

    // Wait for frame update
    <-timer.C

    // World loop iterates
  }
}



// TODO: an UnregisterWeb func if necessary

// RegisterWeb - If not already registered, allocate a channel for web output and true OK.
func (w *World) RegisterWeb() (chan Message, bool) {
  // Check for invalid world or previous allocation
  if w == nil || w.webChan != nil{
    return nil, false
  }

  // Allocate new channel for registered web output
  w.webChan = make(chan Message)
  return w.webChan, true
}


func (w *World) updateStopLights() {
	for idx, stopLight := range w.trafficInfo.stopLights {
		if time.Now().After(stopLight.alarm) {
			for direction, lightState := range w.trafficInfo.stopLights[idx].lightstates {
				if lightState == Green {
					w.trafficInfo.stopLights[idx].lightstates[direction] = Orange //TODO: Maybe switch this to orange too?
					w.trafficInfo.stopLights[idx].alarm = time.Now().Add(time.Second)
					break;
				} else if lightState == Orange {
					w.trafficInfo.stopLights[idx].lightstates[direction] = Red
					w.trafficInfo.stopLights[idx].lightstates[(direction+1)%NumberOfDirections] = Green
					w.trafficInfo.stopLights[idx].alarm = time.Now().Add(time.Second * 5)
					break;
				}
			}
			w.webChan <- Message{
				Type:"Stoplight",
				ID:strconv.Itoa(int(idx)),
				West:strconv.Itoa(int(w.trafficInfo.stopLights[idx].lightstates[West])),
				South:strconv.Itoa(int(w.trafficInfo.stopLights[idx].lightstates[South])),
				East:strconv.Itoa(int(w.trafficInfo.stopLights[idx].lightstates[East])),
				North:strconv.Itoa(int(w.trafficInfo.stopLights[idx].lightstates[North])),
			}
		}
	}
}
