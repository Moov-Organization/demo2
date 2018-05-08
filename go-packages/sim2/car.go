package sim2

import (
  "log"
  "fmt"
)

// car - Describes routine hooks and logic for Cars within a World simulation

// Car - struct for all info needed to manage a Car within a World simulation.
type Car struct {
  // TODO determine if Car needs any additional/public members
  id           uint
  path         Path
  world        *World
  syncChan     chan bool
  sendChan     *chan CarInfo
  ethApi       EthApiInterface
  requestState RequestState
}

type Path struct {
  pickUp Location
  dropOff Location
  currentPos Coords
  currentDir Coords
  currentEdgeId uint
  pathState PathState
  routeEdgeIds []uint
  riderAddress string
}

type Location struct {
  intersect Coords
  edgeID uint
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

type changeDestinationFunction func(*Car)

// NewCar - Construct a new valid Car object
func NewCar(id uint, w *World, sync chan bool, send *chan CarInfo, mrmAddress string, privateKeyString string) *Car {
  c := new(Car)
  c.id = id
  c.world = w
  c.syncChan = sync
  c.sendChan = send
  c.requestState = None

  c.path.pathState = DrivingAtRandom
  edge := c.world.Graph.getRandomEdge()
  c.path.currentPos = edge.Start.Pos
  c.path.currentDir = edge.Start.Pos.UnitVector(edge.End.Pos)
  c.path.currentEdgeId = edge.ID
  c.path.routeEdgeIds, _ = c.getShortestPathToEdge(c.world.Graph.getRandomEdge().ID)

  c.ethApi = NewEthApi(mrmAddress, privateKeyString)
  return c
}

// CarLoop - Begin the car simulation execution loop
func (c *Car) CarLoop() {

  for {
    <-c.syncChan // Block waiting for next sync event
    switch c.path.pathState {
    case DrivingAtRandom:
      c.checkRequestState()
      c.driveToDestination(func (c *Car) {
        c.path.routeEdgeIds, _ = c.getShortestPathToEdge(c.world.Graph.getRandomEdge().ID)
        fmt.Println("To Another Random")
      })
    case ToPickUp:
      c.driveToDestination(func (c *Car) {
        c.path.routeEdgeIds, _ = c.getShortestPathToEdge(c.path.dropOff.edgeID)
        fmt.Println("To Drop off")
        c.path.pathState = ToDropOff
      })
    case ToDropOff:
      c.driveToDestination(func (c *Car) {
        fmt.Println("Back To Random")
        c.path.routeEdgeIds, _ = c.getShortestPathToEdge(c.world.Graph.getRandomEdge().ID)
        c.path.pathState = DrivingAtRandom
      })
    }

    info := CarInfo{ Pos:c.path.currentPos, Vel:Coords{0,0}, Dir:c.path.currentDir }
    *c.sendChan <- info
  }
}

func (c *Car) checkRequestState() {
  if c.requestState != Trying {
    if c.requestState == Fail || c.requestState == None {
      if available, address := c.ethApi.GetRideAddressIfAvailable(); available == true {
        fmt.Println("Found a Ride")
        go c.tryToAcceptRequest(address)
        c.requestState = Trying;
      }
    }else if c.requestState == Success {
      fmt.Println("Got the Ride")
      c.path.pickUp, c.path.dropOff = c.getLocations()
      c.path.routeEdgeIds, _ = c.getShortestPathToEdge(c.path.pickUp.edgeID)
      c.path.pathState = ToPickUp
      c.requestState = None
    }
  }
}

func (c *Car) tryToAcceptRequest(address string) {
  if c.ethApi.AcceptRequest(address) {
    log.Printf("Accept Request success \n")
    c.path.riderAddress = address
    c.requestState = Success
  } else {
    log.Printf("Accept Request failed \n")
    c.requestState = Fail
  }
}

func (c *Car) getLocations() (pickup Location, dropOff Location) {
  from, to := c.ethApi.GetLocations(c.path.riderAddress)
  fmt.Println("locations ",from," ", to)
  x, y := splitCSV(from)
  pickup = c.world.Graph.closestEdgeAndCoord(Coords{x,y})
  x, y = splitCSV(to)
  dropOff = c.world.Graph.closestEdgeAndCoord(Coords{x,y})
  return
}

func (c *Car) getShortestPathToEdge(edgeId uint) (edgeIDs []uint, dist float64) {
  return c.world.Graph.ShortestPath(c.world.Graph.Edges[c.path.currentEdgeId].End.ID, c.world.Graph.Edges[edgeId].Start.ID)
}

func (c *Car) driveToDestination(changeDestination changeDestinationFunction) {
  currentEdge := c.world.Graph.Edges[c.path.currentEdgeId]
  if c.path.currentPos.Equals(currentEdge.End.Pos) {  // Already at edge end, so change edge
    c.path.currentEdgeId = c.path.routeEdgeIds[0]
    currentEdge := c.world.Graph.Edges[c.path.currentEdgeId]
    c.path.currentDir = currentEdge.Start.Pos.UnitVector(currentEdge.End.Pos)
    fmt.Println("next point route ", c.world.Graph.Edges[c.path.routeEdgeIds[0]].End.ID)
    c.path.routeEdgeIds = c.path.routeEdgeIds[1:]
    if len(c.path.routeEdgeIds) == 0 {
      changeDestination(c)
    }
  }else if c.path.currentPos.Distance(currentEdge.End.Pos) > 1.0 {
    c.path.currentPos = c.path.currentPos.ProjectInDirection(1.0, currentEdge.End.Pos)
    //TODO: check for collision
  } else {
    c.path.currentPos = currentEdge.End.Pos
    //TODO: check for stop sign and stop light
  }
}

