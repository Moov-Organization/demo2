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
  syncChan     chan []CarInfo
  sendChan     *chan CarInfo
  ethApi       BlockchainInterface
  requestState RequestState
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
	otherCarInfo			 []CarInfo
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
	stoppedCars   []CarInfo
	movingCars    []CarInfo
	//lastCarSince  time.Time
}


// NewCar - Construct a new valid Car object
func NewCar(id uint, graph *Digraph, ethApi BlockchainInterface, sync chan []CarInfo, send *chan CarInfo) *Car {
  c := new(Car)
  c.id = id
	c.graph = graph
  c.syncChan = sync
  c.sendChan = send
	c.ethApi = ethApi
	if id == 0 {
		id = 27
	}
	if id == 1 {
		id = 20
	}
	if id == 2 {
		id = 5
	}
	if id == 3 {
		id = 14
	}

  c.path.pos = c.graph.Vertices[id].Pos
  c.path.edge = c.graph.Vertices[id].AdjEdges[0]
  c.path.routeEdges, _ = c.getShortestPathToEdge(c.graph.getRandomEdge())
  return c
}

func (c *Car) getShortestPathToEdge(edge Edge) (edges []Edge, dist float64) {
	edges, dist = c.graph.shortestPath(c.path.edge.End.ID, edge.Start.ID)
	//fmt.Print("Car ", c.id," new path ")
	//for _ , edge := range edges {
	//	fmt.Print(edge.End.ID," ")
	//}
	//fmt.Println()
	return
}

// CarLoop - Begin the car simulation execution loop
func (c *Car) CarLoop() {
  for {
  	//TODO pull out the current car's id in the next line
    c.path.otherCarInfo =  <-c.syncChan // Block waiting for next sync event
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
		//	case <-c.path.stopAlarm:
		//		c.allClearToCrossIntersection()
		//	default:
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
	if c.path.destinationEdgeNotReached() {
		c.keepDrivingOnRoute()
	} else {
		switch c.path.state {
		case DrivingAtRandom:
			c.path.routeEdges, _ = c.getShortestPathToEdge(c.graph.getRandomEdge())
			c.path.state = DrivingAtRandom
		case ToPickUp:
			//if c.driveOnCurrentEdgeTowards(c.path.pickUp.intersect) == Reached {
			if c.driveOnCurrentEdgeTowards(c.path.edge.End.Pos)  {
				fmt.Println("Car",c.id," Reached Pick Up, To Drop off")
				c.path.routeEdges, _ = c.getShortestPathToEdge(c.path.dropOff.edge)
				c.path.state = Waiting
				c.path.nextState = ToDropOff
				c.path.stopAlarm = time.After(time.Second * 1)
			}
		case ToDropOff:
			//if c.driveOnCurrentEdgeTowards(c.path.dropOff.intersect) == Reached {
			if c.driveOnCurrentEdgeTowards(c.path.edge.End.Pos) {
				fmt.Println("Car",c.id," Reached Drop off, back to Random")
				c.path.routeEdges, _ = c.getShortestPathToEdge(c.graph.getRandomEdge())
				c.path.state = Waiting
				c.path.nextState = DrivingAtRandom
				c.path.stopAlarm = time.After(time.Second * 2)
			}
		}
	}
}

func (p *Path) destinationEdgeNotReached() (bool) {
	return len(p.routeEdges) != 1
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
				//if c.clearToPassStopLight() {
				//	c.path.loadNextEdge()
				//}
			}
		} else {
			c.path.loadNextEdge()
		}
	}	else if c.path.pos.Distance(c.path.edge.End.Pos) > MovementPerFrame {
		if !c.collisionAhead() {
			c.path.pos = c.path.pos.ProjectInDirection(MovementPerFrame, c.path.edge.End.Pos)
		}
	} else {
		c.path.pos = c.path.edge.End.Pos
		c.path.justReachedEdgeEnd = true
	}
}

func (c *Car) saveStopSignInfo() {
	otherCars := removeCar(c.path.otherCarInfo, c.id)
	c.path.waitingFor.movingCars = c.getCarsMovingInIntersection(otherCars)
	c.path.waitingFor.stoppedCars = c.getCarsStoppedAtIntersection(otherCars)
}

func (c *Car) clearToPassStopSign() (clear bool) {
	clear = true
	otherCars := removeCar(c.path.otherCarInfo, c.id)

	currentlyMovingCars := c.getCarsMovingInIntersection(otherCars)
	for _, alreadyMovingCar := range c.path.waitingFor.movingCars {
		if carIsPresent(currentlyMovingCars, alreadyMovingCar.ID) {
			fmt.Println("Car ", c.id," waiting on car ", alreadyMovingCar.ID," to finish crossing")
			clear = false
		} else {
			c.path.waitingFor.movingCars = removeCar(c.path.waitingFor.movingCars, alreadyMovingCar.ID)
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

	//TODO break deadlocks if cars arrive at the same time
	//if len(c.path.waitingFor.movingCars) == 0 && len(c.path.waitingFor.stoppedCars) == 1 {
	//	if time.Now().Sub(c.path.waitingFor.lastCarSince) > time.Second * 10 {
	//		//c.path.waitingFor.stoppedCars = []CarInfo{}
	//		//clear = true
	//		fmt.Println("HERE")
	//	}
	//}
	return
}


func (c *Car) collisionAhead() bool {
	carInfos := removeCar(c.path.otherCarInfo, c.id)
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
	numbers = splitLine(to, ",", 2)
	dropOff = c.graph.closestEdgeAndCoord(Coords{numbers[0],numbers[1]})
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
					}else if edge.End.AdjEdges[0].End.ID >= 32 && edge.End.AdjEdges[0].End.ID == otherCarInfo.EdgeId { //Hack to include edge
						movingCars = append(movingCars, otherCarInfo)
					}
				}
			}
		}
	}
	return
}
