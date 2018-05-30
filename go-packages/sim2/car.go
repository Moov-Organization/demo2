package sim2

import (
  "log"
  "fmt"
  "time"
)

// car - Describes routine hooks and logic for Cars within a World simulation

// The distance the car should move every drive call
const MovementPerFrame = 2

// Car - struct for all info needed to manage a Car within a World simulation.
type Car struct {
  // TODO determine if Car needs any additional/public members
  id           uint
  path         Path
  graph        *Digraph
  syncChan     chan TrafficInfo
  sendChan     *chan CarInfo
  ethApi       BlockchainInterface
  requestState RequestState
	webChan chan Message
}

type Path struct {
  pickUp             Location
  dropOff            Location
  pos                Coords
  edge               Edge
  state              PathState
  nextState          PathState
  routeEdges         []Edge
  riderAddress       string
  justReachedEdgeEnd bool
  stopAlarm          <-chan time.Time
  waitingFor         IntersectionContext
  trafficInfo       TrafficInfo
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

type IntersectionContext struct {
	stoppedCars      []CarInfo
	movingCars       []CarInfo
	noCarsMoveSince  time.Time
}


// NewCar - Construct a new valid Car object
func NewCar(id uint, graph *Digraph, ethApi BlockchainInterface, sync chan TrafficInfo, send *chan CarInfo, webChan chan Message) *Car {
  c := new(Car)
  c.id = id
  c.graph = graph
  c.syncChan = sync
  c.sendChan = send
  c.ethApi = ethApi
  c.webChan = webChan
  c.path.pos = c.graph.Vertices[id].Pos
  c.path.edge = *c.graph.Vertices[id].AdjEdges[0]
  //TODO deal with random edge equals current edge
  c.path.routeEdges, _ = c.getShortestPathToEdge(c.graph.getRandomEdge())

  return c
}

func (c *Car) getShortestPathToEdge(edge Edge) (edges []Edge, dist float64) {
	edges, dist = c.graph.shortestPath(c.path.edge.End.ID, edge.Start.ID)
	edges = append(edges, edge)
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
    //TODO pull out the current car's id in the next line
    c.path.trafficInfo =  <-c.syncChan // Block waiting for next sync event
    c.drive()
    info := CarInfo{ID:c.id, Pos:c.path.pos, Vel:Coords{0,0}, Dir:c.path.edge.unitVector(), EdgeId:c.path.edge.ID }
    *c.sendChan <- info
  }
}

func (c *Car) drive () {
  switch c.path.state {
  case DrivingAtRandom:
    c.checkRequestState()
    c.driveToDestination()
  case ToPickUp:
    c.driveToDestination()
  case ToDropOff:
    c.driveToDestination()
  case Stopped:
    //select {
    //  case <-c.path.stopAlarm:
    //    c.allClearToCrossIntersection()
    //  default:
    //}
  case Waiting:
    select {
    case <-c.path.stopAlarm:
      c.path.state = c.path.nextState
    default:
    }

  }
}

func (c *Car) driveToDestination() {
  if !c.path.destinationEdgeReached() {
    c.keepDrivingOnRoute()
  } else {
    switch c.path.state {
    case DrivingAtRandom:
      c.path.routeEdges, _ = c.getShortestPathToEdge(c.graph.getRandomEdge())
      c.path.state = DrivingAtRandom
    case ToPickUp:
      if c.driveOnCurrentEdgeTowards(c.path.pickUp.intersect) {
        fmt.Println("Car",c.id," Reached Pick Up, To Drop off")
				c.webChan <- Message{
					Type:"RideStatus",
					Address:c.path.riderAddress,
					State:"At Pick Up",
				}
        c.path.routeEdges, _ = c.getShortestPathToEdge(c.path.dropOff.edge)
        c.path.state = Waiting
        c.path.nextState = ToDropOff
        c.path.stopAlarm = time.After(time.Second * 5)
      }
    case ToDropOff:
      if c.driveOnCurrentEdgeTowards(c.path.dropOff.intersect) {
        fmt.Println("Car",c.id," Reached Drop off, back to Random")
        c.webChan <- Message{
					Type:"RideStatus",
					Address:c.path.riderAddress,
					State:"At Drop Off",
				}
        c.path.routeEdges, _ = c.getShortestPathToEdge(c.graph.getRandomEdge())
        c.path.state = Waiting
        c.path.nextState = DrivingAtRandom
        c.path.stopAlarm = time.After(time.Second * 5)
      }
    }
  }
}

func (p *Path) destinationEdgeReached() (bool) {
  return len(p.routeEdges) == 0
}

func (p *Path) loadNextEdge() () {
  p.edge = p.routeEdges[0]
  p.routeEdges = p.routeEdges[1:]
  p.justReachedEdgeEnd = false;
}

func (c *Car) driveOnCurrentEdgeTowards(endPos Coords) (bool) {
  if c.path.pos == endPos {
    return true
  }else if c.path.pos.Distance(endPos) < MovementPerFrame {
    c.path.pos = endPos
    return true
  } else {
    if !c.collisionAhead() {
      c.path.pos = c.path.pos.ProjectInDirection(MovementPerFrame, c.path.edge.End.Pos)
    }
    return false
  }
}

func (c *Car) keepDrivingOnRoute() () {

  if c.path.pos == c.path.edge.End.Pos {
    if c.path.edge.End.intersection != nil {
      switch c.path.edge.End.intersection.intersectionType {
      case StopSign:
        if c.path.justReachedEdgeEnd { // Just reached waitingFor
          //fmt.Println("Car ", c.id," Reached Stop Sign")
          c.saveStopSignInfo()
          c.path.nextState = c.path.state
          c.path.state = Waiting
          c.path.stopAlarm = time.After(time.Second * 2)
          c.path.justReachedEdgeEnd = false
        } else if c.clearToPassStopSign() {
          //fmt.Println("Car ", c.id," clear to cross intersection")
          c.path.loadNextEdge()
        } else {
          c.path.nextState = c.path.state
          c.path.state = Waiting
          c.path.stopAlarm = time.After(time.Millisecond * 500)
        }

      case StopLight:
        if c.clearToPassStopLight() {
         c.path.loadNextEdge()
        } else {
					c.path.nextState = c.path.state
					c.path.state = Waiting
					c.path.stopAlarm = time.After(time.Millisecond * 500)
				}
      }
    } else {
      c.path.loadNextEdge()
    }
  } else if c.path.pos.Distance(c.path.edge.End.Pos) > MovementPerFrame {
    if !c.collisionAhead() {
      c.path.pos = c.path.pos.ProjectInDirection(MovementPerFrame, c.path.edge.End.Pos)
    }
  } else {
    c.path.pos = c.path.edge.End.Pos
    c.path.justReachedEdgeEnd = true
  }
}

func (c *Car) saveStopSignInfo() {
	otherCars := removeCar(c.path.trafficInfo.carStates, c.id)
	c.path.waitingFor.movingCars = c.getCarsMovingInIntersection(otherCars)
	c.path.waitingFor.stoppedCars = c.getCarsStoppedAtIntersection(otherCars)
	if len(c.path.waitingFor.movingCars) == 0 {
		c.path.waitingFor.noCarsMoveSince = time.Now()
	}
}

func (c *Car) clearToPassStopSign() (clear bool) {
  clear = true
  otherCars := removeCar(c.path.trafficInfo.carStates, c.id)

	currentlyMovingCars := c.getCarsMovingInIntersection(otherCars)
	for _, alreadyMovingCar := range c.path.waitingFor.movingCars {
		if carIsPresent(currentlyMovingCars, alreadyMovingCar.ID) {
			fmt.Println("Car ", c.id," waiting on car ", alreadyMovingCar.ID," to finish crossing")
			clear = false
		} else {
			c.path.waitingFor.movingCars = removeCar(c.path.waitingFor.movingCars, alreadyMovingCar.ID)
			if len(c.path.waitingFor.movingCars) == 0 {
				c.path.waitingFor.noCarsMoveSince = time.Now()
			}
		}
	}

  currentlyStoppedCars := c.getCarsStoppedAtIntersection(otherCars)
  for _, alreadyStoppedCar := range c.path.waitingFor.stoppedCars {
    if carIsPresent(currentlyStoppedCars, alreadyStoppedCar.ID) {
      fmt.Println("Car ", c.id," waiting on car ", alreadyStoppedCar.ID," to start crossing ")
      clear = false
    } else {
      c.path.waitingFor.stoppedCars = removeCar(c.path.waitingFor.stoppedCars, alreadyStoppedCar.ID)
      if carIsPresent(currentlyMovingCars, alreadyStoppedCar.ID) {
        fmt.Println("Car ", c.id," waiting on car ", alreadyStoppedCar.ID," to finish crossing")
        c.path.waitingFor.movingCars = append(c.path.waitingFor.movingCars, alreadyStoppedCar)
        clear = false
      } else {
        // Good to go
      }
    }
  }

	// break deadlocks if cars arrive at the same time
	if len(c.path.waitingFor.movingCars) == 0 && len(c.path.waitingFor.stoppedCars) > 0 {
		if time.Now().Sub(c.path.waitingFor.noCarsMoveSince) > time.Second * 5 {
			clear = true
			for _, stoppedCar := range c.path.waitingFor.stoppedCars {
				if stoppedCar.ID > c.id {
					clear = false
					fmt.Println("Broke Dead Lock")
				}
			}
		}
	}
	return
}


func (c *Car) collisionAhead() bool {
  carInfos := removeCar(c.path.trafficInfo.carStates, c.id)
  //fmt.Printf("car %d info address %p \n", c.id, &carInfos)
  for _, otherCarInfo := range carInfos {
    if otherCarInfo.ID != c.id && otherCarInfo.EdgeId == c.path.edge.ID {
      edgeEndPos := c.path.edge.End.Pos
      otherCarDistanceToEdge := otherCarInfo.Pos.Distance(edgeEndPos)
      thisCarDistanceToEdge := c.path.pos.Distance(edgeEndPos)
      distanceBetweenCars := thisCarDistanceToEdge - otherCarDistanceToEdge
      if distanceBetweenCars > 0 && distanceBetweenCars < 100 {
        return true
      }
    }
  }
  return false
}

func (c *Car) clearToPassStopLight() (clear bool) {
	for _, stopLight := range c.path.trafficInfo.stopLightStates {
		if stopLight.ID == c.path.edge.End.intersection.id {
			entries := c.path.edge.End.intersection.entries
			for direction, intersectionEntry := range entries {
				if intersectionEntry.present && intersectionEntry.vertex.Pos == c.path.pos{
					if stopLight.lightstates[direction] == Green && len(c.getCarsMovingInIntersection(c.path.trafficInfo.carStates)) == 0 {
						return true
					} else {
						return false
					}
				}
			}
		}
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
			c.webChan <- Message{
				Type:"RideStatus",
				Address:c.path.riderAddress,
				State:"To Pick Up",
			}
      c.path.pickUp, c.path.dropOff = c.getLocations()
      c.path.routeEdges, _ = c.getShortestPathToEdge(c.path.pickUp.edge)
      c.path.state = ToPickUp
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
  //TODO: recover from improperly formatted locations
  from, to := c.ethApi.GetLocations(c.path.riderAddress)
  fmt.Println("Car",c.id," locations ",from," ", to)
  numbers := splitLine(from, ",", 2)
  pickup = c.graph.closestEdgeAndCoord(Coords{numbers[0],numbers[1]})
  fmt.Println("Pick Up", pickup.edge.Start.ID, " ", pickup.edge.End.ID)
  numbers = splitLine(to, ",", 2)
  dropOff = c.graph.closestEdgeAndCoord(Coords{numbers[0],numbers[1]})
  fmt.Println("Drop Off", dropOff.edge.Start.ID, " ", dropOff.edge.End.ID)
  return
}

func removeCar(cars []CarInfo, blackSheep uint) ([]CarInfo) {
  for i, car := range cars {
    if car.ID == blackSheep {
      return append(cars[:i], cars[i+1:]...)
    }
  }
  return cars
}

func carIsPresent(cars []CarInfo, blackSheep uint) (bool) {
  for _, car := range cars {
    if car.ID == blackSheep {
      return true
    }
  }
  return false
}

func (c *Car) getCarsStoppedAtIntersection(otherCars []CarInfo) (stoppedCars []CarInfo) {
  for _, otherCarInfo := range otherCars {
    entries := c.path.edge.End.intersection.entries
    for _, intersectionEntry := range entries {
      if intersectionEntry.present && intersectionEntry.vertex.Pos == otherCarInfo.Pos {
        stoppedCars = append(stoppedCars, otherCarInfo)
      }
    }
  }
  return
}

func (c *Car) getCarsMovingInIntersection(otherCars []CarInfo) (movingCars []CarInfo) {
  for _, otherCarInfo := range otherCars {
    entries := c.path.edge.End.intersection.entries
    for _, intersectionEntry := range entries {
      if intersectionEntry.present {
        for _, edge := range intersectionEntry.vertex.AdjEdges {
          if edge.ID == otherCarInfo.EdgeId {
            movingCars = append(movingCars, otherCarInfo)
          }else if edge.Extends && edge.End.AdjEdges[0].ID == otherCarInfo.EdgeId {
            movingCars = append(movingCars, otherCarInfo)
          }
        }
      }
    }
  }
  return
}
