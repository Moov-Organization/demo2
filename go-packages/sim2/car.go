package sim2

import (
  "log"
  "fmt"
	"time"
)

// car - Describes routine hooks and logic for Cars within a World simulation

// The distance the car should move every drive call
const MovementPerFrame = 0.5

// Car - struct for all info needed to manage a Car within a World simulation.
type Car struct {
  // TODO determine if Car needs any additional/public members
  id           uint
  path         Path
  world        CarWorldInterface
  syncChan     chan bool
  sendChan     *chan CarInfo
  ethApi       BlockchainInterface
  requestState RequestState
}

type Path struct {
  pickUp       Location
  dropOff      Location
  currentPos   Coords
  currentEdge  Edge
  currentState PathState
	nextState    PathState
  routeEdges   []Edge
  riderAddress string
  reachedEdgeEnd bool
	stopAlarm    <-chan time.Time
}

type Location struct {
  intersect Coords
  edge Edge
}

type PathState int
const (
  DrivingAtRandom   PathState = 0
  ToPickUp          PathState = 1
  ToDropOff         PathState = 2
  Stopped           PathState = 3
  Waiting           PathState = 4
)

type RequestState int
const (
  None    RequestState = 0
  Trying  RequestState = 1
  Success RequestState = 2
  Fail    RequestState = 3
)

type ReachedState int
const (
  NotReached ReachedState = 0
  Reached    ReachedState = 1
)

// NewCar - Construct a new valid Car object
func NewCar(id uint, w CarWorldInterface, ethApi BlockchainInterface, sync chan bool, send *chan CarInfo) *Car {
  c := new(Car)
  c.id = id
  c.world = w
  c.syncChan = sync
  c.sendChan = send
	c.ethApi = ethApi
	if id == 1 {
		id = 2
	}
  c.path.currentPos = c.world.getVertex(id).Pos
  c.path.currentEdge = c.world.getVertex(id).AdjEdges[0]
  c.path.routeEdges, _ = c.getShortestPathToEdge(w.getRandomEdge())
  return c
}

func (c *Car) getShortestPathToEdge(edge Edge) (edges []Edge, dist float64) {
	edges, dist = c.world.shortestPath(c.path.currentEdge.End.ID, edge.Start.ID)
	fmt.Print("Car ", c.id," new path ")
	for _ , edge := range edges {
		fmt.Print(edge.End.ID," ")
	}
	fmt.Println()
	return
}

// CarLoop - Begin the car simulation execution loop
func (c *Car) CarLoop() {
  for {
    <-c.syncChan // Block waiting for next sync event
    c.drive()
    info := CarInfo{ID:c.id, Pos:c.path.currentPos, Vel:Coords{0,0}, Dir:c.path.currentEdge.unitVector(), EdgeId:c.path.currentEdge.ID }
    *c.sendChan <- info
  }
}

func (c *Car) drive () {
  switch c.path.currentState {
  case DrivingAtRandom:
    c.checkRequestState()
  	c.driveToDestination()
  case ToPickUp:
		c.driveToDestination()
  case ToDropOff:
		c.driveToDestination()
	case Stopped:
		select {
			case <-c.path.stopAlarm:
				c.path.currentState = c.path.nextState
			default:
		}
  case Waiting:
		select {
		case <-c.path.stopAlarm:
			c.path.currentState = c.path.nextState
		default:
		}

  }
}

func (c *Car) driveToDestination() {
	if !c.path.atLastEdge() {
		if c.driveOnCurrentEdgeTowards(c.path.currentEdge.End.Pos) == Reached {
			if c.allClearToSwitchToNextEdge() {
				c.path.currentEdge = c.path.routeEdges[0]
				c.path.routeEdges = c.path.routeEdges[1:]
				fmt.Println("Car",c.id," To ",c.path.currentEdge.End.ID)
			}
		}
	} else {
		switch c.path.currentState {
		case DrivingAtRandom:
			c.path.routeEdges, _ = c.getShortestPathToEdge(c.world.getRandomEdge())
			c.path.currentState = DrivingAtRandom
		case ToPickUp:
			//if c.driveOnCurrentEdgeTowards(c.path.pickUp.intersect) == Reached {
			if c.driveOnCurrentEdgeTowards(c.path.currentEdge.End.Pos) == Reached {
				fmt.Println("Car",c.id," Reached Pick Up, To Drop off")
				c.path.routeEdges, _ = c.getShortestPathToEdge(c.path.dropOff.edge)
				c.path.currentState = Waiting
				c.path.nextState = ToDropOff
				c.path.stopAlarm = time.After(time.Second * 2)
				c.path.reachedEdgeEnd = false
			}
		case ToDropOff:
			//if c.driveOnCurrentEdgeTowards(c.path.dropOff.intersect) == Reached {
			if c.driveOnCurrentEdgeTowards(c.path.currentEdge.End.Pos) == Reached {
				fmt.Println("Car",c.id," Reached Drop off, back to Random")
				c.path.routeEdges, _ = c.getShortestPathToEdge(c.world.getRandomEdge())
				c.path.currentState = Waiting
				c.path.nextState = DrivingAtRandom
				c.path.stopAlarm = time.After(time.Second * 2)
				c.path.reachedEdgeEnd = false
			}
		}
	}
}

func (p *Path) atLastEdge() (bool) {
	return len(p.routeEdges) == 1
}

func (c *Car) driveOnCurrentEdgeTowards(endPos Coords) (ReachedState) {
	if c.path.currentPos == endPos {
		return Reached
	}
	if c.path.currentPos.Distance(endPos) > MovementPerFrame {
		if !c.collisionAhead() {
			c.path.currentPos = c.path.currentPos.ProjectInDirection(MovementPerFrame, c.path.currentEdge.End.Pos)
		}
		return NotReached
	} else {
		c.path.currentPos = endPos
		c.path.reachedEdgeEnd = true
		return Reached
	}
}

func (c *Car) collisionAhead() bool {
	carInfos := c.world.getCarInfos()
	//fmt.Printf("car %d info address %p \n", c.id, &carInfos)
	for otherCarId, otherCarInfo := range carInfos {
		if uint(otherCarId) != c.id && otherCarInfo.EdgeId == c.path.currentEdge.ID {
			edgeEndPos := c.path.currentEdge.End.Pos
			otherCarDistanceToEdge := otherCarInfo.Pos.Distance(edgeEndPos)
			thisCarDistanceToEdge := c.path.currentPos.Distance(edgeEndPos)
			distanceBetweenCars := thisCarDistanceToEdge - otherCarDistanceToEdge
			if distanceBetweenCars > 0 && distanceBetweenCars < 100 {
				return true
			}
		}
	}
	return false
}

func (c *Car) allClearToSwitchToNextEdge() (proceed bool) {
	if c.path.reachedEdgeEnd && len(c.path.currentEdge.End.AdjEdges) > 1{
		c.path.nextState = c.path.currentState
		c.path.currentState = Stopped
		c.path.stopAlarm = time.After(time.Second * 2)
		c.path.reachedEdgeEnd = false
		proceed = false
	} else {
		proceed = true
	}
	return
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