package sim2

import (
  "log"
  "fmt"
)

// car - Describes routine hooks and logic for Cars within a World simulation

// The distance the car should move every drive call
const MovementPerFrame = 1.0

// Car - struct for all info needed to manage a Car within a World simulation.
type Car struct {
  // TODO determine if Car needs any additional/public members
  id           uint
  path         Path
  world        CarWorldInterface
  syncChan     chan bool
  sendChan     *chan CarInfo
  ethApi       EthApiInterface
  requestState RequestState
}

type Path struct {
  pickUp       Location
  dropOff      Location
  currentPos   Coords
  currentDir   Coords
  currentEdge  *Edge
  currentState PathState
  routeEdges   []*Edge
  riderAddress string
}

type Location struct {
  intersect Coords
  edge *Edge
}

type PathState int
const (
  DrivingAtRandom PathState = 0
  ToPickUp        PathState = 1
  ToDropOff       PathState = 2
)

type RequestState int
const (
  None    RequestState = 0
  Trying  RequestState = 1
  Success RequestState = 2
  Fail    RequestState = 3
)

// NewCar - Construct a new valid Car object
func NewCar(id uint, w CarWorldInterface, ethApi EthApiInterface, sync chan bool, send *chan CarInfo) *Car {
  c := new(Car)
  c.id = id
  c.world = w
  c.syncChan = sync
  c.sendChan = send
  c.requestState = None

  c.path.currentState = DrivingAtRandom
  c.path.currentPos = c.world.getVertex(id).Pos
  c.path.currentEdge = c.world.getVertex(id).AdjEdges[0]
  c.path.routeEdges, _ = c.getShortestPathToEdge(w.getRandomEdge())
  c.ethApi = ethApi
  return c
}


// CarLoop - Begin the car simulation execution loop
func (c *Car) CarLoop() {
  for {
    <-c.syncChan // Block waiting for next sync event
    c.drive()
    info := CarInfo{ID:c.id, Pos:c.path.currentPos, Vel:Coords{0,0}, Dir:c.path.currentDir }
    *c.sendChan <- info
  }
}

func (c *Car) checkRequestState() {
  if c.requestState == Trying {
    return
  } else {
    if c.requestState == Fail || c.requestState == None {
      if available, address := c.ethApi.GetRideAddressIfAvailable(); available == true {
        fmt.Println("Car",c.id," Found a Ride")
        go c.tryToAcceptRequest(address)
        c.requestState = Trying;
      }
    }else if c.requestState == Success {
      fmt.Println("Car",c.id," Got the Ride, To Pick Up")
      c.path.pickUp, c.path.dropOff = c.getLocations()
      c.path.routeEdges, _ = c.getShortestPathToEdge(c.path.pickUp.edge)
      c.path.currentState = ToPickUp
      c.requestState = None
    }
  }
}

func (c *Car) tryToAcceptRequest(address string) {
  if c.ethApi.AcceptRequest(address) {
    log.Println("Car ",c.id," Accept Request success")
    c.path.riderAddress = address
    c.requestState = Success
  } else {
    log.Println("Car ",c.id," Accept Request failed")
    c.requestState = Fail
  }
}

func (c *Car) getLocations() (pickup Location, dropOff Location) {
  from, to := c.ethApi.GetLocations(c.path.riderAddress)
  fmt.Println("Car",c.id," locations ",from," ", to)
  x, y := splitCSV(from)
  pickup = c.world.closestEdgeAndCoord(Coords{x,y})
  x, y = splitCSV(to)
  dropOff = c.world.closestEdgeAndCoord(Coords{x,y})
  return
}

func (c *Car) getShortestPathToEdge(edge *Edge) (edges []*Edge, dist float64) {
  return c.world.ShortestPath(c.path.currentEdge.End.ID, edge.Start.ID)
}


func (c *Car) drive() {
  if c.path.currentState == DrivingAtRandom {
	c.checkRequestState()
  }
  if c.path.currentPos.Equals(c.path.currentEdge.End.Pos) {  // Already at edge end, so change edge
    c.path.currentEdge = c.path.routeEdges[0]
    c.path.currentDir = c.path.currentEdge.unitVector()
    //fmt.Println("next point route ", c.world.getEdge(c.path.routeEdgeIds[0]).End.ID)
    // Remove the first element of the queue
    c.path.routeEdges = c.path.routeEdges[1:]
    // If queue is empty than get next destination
    if len(c.path.routeEdges) == 0 {
      if c.path.currentState == DrivingAtRandom || c.path.currentState == ToDropOff{
		  c.path.routeEdges, _ = c.getShortestPathToEdge(c.world.getRandomEdge())
		  fmt.Println("Car",c.id," To Random")
		  c.path.currentState = DrivingAtRandom
	  } else if c.path.currentState == ToPickUp {
		  c.path.routeEdges, _ = c.getShortestPathToEdge(c.path.dropOff.edge)
		  fmt.Println("Car",c.id," To Drop off")
		  c.path.currentState = ToDropOff
	  }
    }
  }else if c.path.currentPos.Distance(c.path.currentEdge.End.Pos) > MovementPerFrame {
  	c.path.currentPos = c.path.currentPos.ProjectInDirection(MovementPerFrame, c.path.currentEdge.End.Pos)
    //TODO: check for collision
  } else {
    c.path.currentPos = c.path.currentEdge.End.Pos
    //TODO: check for stop sign and stop light
  }
  return
}
