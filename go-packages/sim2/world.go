package sim2

import (
  "time"
  "strconv"

)

// world - Describes world state: position and velocity of all cars in simulation

// CarInfo - struct to contain position and velocity information for a simulated car.
type CarInfo struct {
  ID uint
  Pos Coords
  Vel Coords  // with respect to current position, offset for a single frame
  Dir Coords  // unit vector with respect to current position
  EdgeId uint
}

// World - struct to contain all relevat world information in simulation.
type World struct {
  graph *Digraph
  carStates []CarInfo
  fps float64
  numRegisteredCars uint  // Total count of registered cars
  syncChans []chan []CarInfo  // Index by actor ID for channel to/from that actor
  recvChan chan CarInfo  // Receive from all Cars registered on one channel
  webChan chan Message
}

// NewWorld - Constructor for valid World object.
func NewWorld(fps float64, graph *Digraph) *World {
  w := new(World)
  w.graph = graph
  w.fps = fps
  w.numRegisteredCars = 0
  // NOTE recvChan is nil until cars are registered
  // NOTE webChan is nil until registered
  return w
}

// RegisterCar - If the car ID has not been taken, allocate new channels for the car ID and true OK.
//   If the car ID is taken or World is unallocated, return nil channels and false OK value.
func (w *World) RegisterCar() (uint, chan []CarInfo, *chan CarInfo, bool) {
  // Check for invalid World
  if w == nil {
    return 0, nil, nil, false
  }
  ID := w.numRegisteredCars

  w.numRegisteredCars++
  //fmt.Println("numRegisteredCars:", w.numRegisteredCars)

  // Allocate new channels for registered car
  w.carStates = append(w.carStates, CarInfo{ ID:ID }) // TODO: randomize/control car location on startup
  w.syncChans = append(w.syncChans, make(chan []CarInfo, 1)) // Buffer up to one output
  w.recvChan = make(chan CarInfo, w.numRegisteredCars)  // Overwrite buffered allocation
  return ID, w.syncChans[ID], &w.recvChan, true
}

// TODO: an UnregisterCar func if necessary

// LoopWorld - Begin the world simulation execution loop
func (w *World) LoopWorld() {
  itercounter := uint64(0)
  for {
    //fmt.Println("Iteration", itercounter)
    timer := time.NewTimer(time.Duration(1000/w.fps) * time.Millisecond)

    // Send out sync flag = true for each registered car
    for ID := range w.carStates {
      cpy := make([]CarInfo, len(w.carStates))
      copy(cpy, w.carStates)
      w.syncChans[ID] <- cpy
    }

    // Car coroutines should now process current world state
    for idx, car := range w.carStates {
      w.webChan <- Message{
        Type:"Car",
        ID:strconv.Itoa(int(idx)),
        X:strconv.Itoa(int(car.Pos.X)),
        Y:strconv.Itoa(int(car.Pos.Y)),
        Orientation:strconv.Itoa(int(Coords{0,0}.Angle(car.Dir))),
      }
    }

    // Wait for all registered cars to report
    for carRecvCt := uint(0); carRecvCt < w.numRegisteredCars ; {
      data := <-w.recvChan
      // TODO: deep copy is safer here
      w.carStates[data.ID] = data

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


