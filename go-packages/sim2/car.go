package sim2

import (
  "math"
  "log"
  "math/rand"
  "time"
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
  WaitingForRequest PathState = 0
  ToPickUp PathState = 1
  ToDropOff PathState = 2
)

type RequestState int
const (
  None    RequestState = 0
  Trying  RequestState = 1
  Success RequestState = 2
  Fail    RequestState = 3
)

// NewCar - Construct a new valid Car object
func NewCar(id uint, w *World, sync chan bool, send *chan CarInfo, mrmAddress string, privateKeyString string) *Car {
  c := new(Car)
  c.id = id
  c.world = w
  c.syncChan = sync
  c.sendChan = send
  c.requestState = None


  s1 := rand.NewSource(time.Now().UnixNano())
  r1 := rand.New(s1)
  edge := c.world.Graph.Edges[uint(r1.Int() % len(c.world.Graph.Edges))]

  c.path.pathState = WaitingForRequest
  c.path.currentPos = edge.Start.Pos
  c.path.currentDir = edge.Start.Pos.UnitVector(edge.End.Pos)
  c.path.currentEdgeId = edge.ID

  c.ethApi = NewEthApi(mrmAddress, privateKeyString)
  return c
}

// CarLoop - Begin the car simulation execution loop
func (c *Car) CarLoop() {

  for {
      <-c.syncChan // Block waiting for next sync event

      if c.path.pathState == WaitingForRequest {
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

      if c.path.pathState == ToPickUp || c.path.pathState == ToDropOff {
        currentEdge := c.world.Graph.Edges[c.path.currentEdgeId]
        if c.path.currentPos.Equals(currentEdge.End.Pos) {  // Already at edge end, so change edge
          c.path.currentEdgeId = c.path.routeEdgeIds[0]
          currentEdge := c.world.Graph.Edges[c.path.currentEdgeId]
          c.path.currentDir = currentEdge.Start.Pos.UnitVector(currentEdge.End.Pos)
          fmt.Println("next point route ", c.world.Graph.Edges[c.path.routeEdgeIds[0]].End.ID)
          c.path.routeEdgeIds = c.path.routeEdgeIds[1:]
          if len(c.path.routeEdgeIds) == 0 {
            if c.path.pathState == ToPickUp {
              c.path.routeEdgeIds, _ = c.getShortestPathToEdge(c.path.dropOff.edgeID)
              fmt.Println("To Drop off")
              c.path.pathState = ToDropOff
            } else {
              fmt.Println("Back To Random")
              c.path.pathState = WaitingForRequest
            }
          }
        }else if c.path.currentPos.Distance(currentEdge.End.Pos) > 1.0 {
          c.path.currentPos = c.path.currentPos.ProjectInDirection(1.0, currentEdge.End.Pos)
        } else {
          c.path.currentPos = currentEdge.End.Pos
        }
      } else if c.path.pathState == WaitingForRequest {
        endVertex := c.world.Graph.Edges[c.path.currentEdgeId].End
        if c.path.currentPos.Equals(endVertex.Pos) {  // Already at edge end, so change edge
          if len(endVertex.AdjEdges) == 0 {fmt.Println(endVertex.ID)}
          c.path.currentEdgeId = endVertex.AdjEdges[0].ID
          c.path.currentDir = endVertex.AdjEdges[0].Start.Pos.UnitVector(endVertex.AdjEdges[0].End.Pos)
        }else if c.path.currentPos.Distance(endVertex.Pos) > 1.0 {
          c.path.currentPos = c.path.currentPos.ProjectInDirection(1.0, endVertex.Pos)
        } else {
          c.path.currentPos = endVertex.Pos
        }
      }
      // TODO: this

      // Evaluate destination based on current location and status of bids
      // TODO: this

      // Check for shortest path to destination
      // TODO: this

      // Evaluate road rules and desired movement within world
      // TODO: this

      // Send movement update request to World
      // TODO: replace this with real update
      info := CarInfo{ Pos:c.path.currentPos, Vel:Coords{0,0}, Dir:c.path.currentDir }
      *c.sendChan <- info
      //fmt.Println("Car", c.id, ": sent update", inf)
  }
}

func (c *Car) tryToAcceptRequest(address string) {
  if c.ethApi.AcceptRequest(address) {
    log.Printf("Accept Request failed \n")
    c.requestState = Fail
    fmt.Println("Transaction Failed")
  } else {
    log.Printf("Accept Request success \n")
    c.path.riderAddress = address
    c.requestState = Success
    fmt.Println("Transaction Success")
  }
}

func (c *Car) getLocations() (pickup Location, dropOff Location) {
  from, to := c.ethApi.GetLocations(c.path.riderAddress)
  fmt.Println("locations ",from," ", to)
  x, y := splitCSV(from)
  pickup = c.world.closestEdgeAndCoord(Coords{x,y})
  x, y = splitCSV(to)
  dropOff = c.world.closestEdgeAndCoord(Coords{x,y})
  return
}

func (c *Car) getShortestPathToEdge(edgeId uint) (edgeIDs []uint, dist float64) {
  return c.world.Graph.ShortestPath(c.world.Graph.Edges[c.path.currentEdgeId].End.ID, c.world.Graph.Edges[edgeId].Start.ID)
}

// closestEdgeAndCoord For coords within world space, find  closest coords on an edge on world graph
// Return coordinates of closest point on world graph, and corresponding edge ID in world struct
func (w *World) closestEdgeAndCoord(queryPoint Coords) (location Location) {
  // TODO: input sanitation/validation; error handling?
  // TODO: proper helper function breakdown of closestEdgeAndCoord

  shortestDistance := math.Inf(1)
  location.intersect = Coords{0, 0}
  location.edgeID = 0

  // TODO: remove randomness caused by traversing equivalent closest edges with 'range' on map here
  for edgeIdx, edge := range w.Graph.Edges {
    coord, dist := edge.checkIntersect(queryPoint)
    //fmt.Print("[", edgeIdx, "] <ClosestEdgeAndCoord>")
    //fmt.Print(" query: ", queryPoint, ", shortest: ", shortestDistance, ", new: ", coord, ", dist: ", dist)
    if dist < shortestDistance {
      shortestDistance = dist
      location.intersect = coord
      location.edgeID = edgeIdx
      //fmt.Print(" (new shortest: ", dist , " @", coord, ")")
    }
    //fmt.Println()
  }
  return
}

// checkIntersect find the point on this edge closest to the query coordinates and report distance
// Return coordinates of closest point and distance of closest point to query
func (e *Edge) checkIntersect(query Coords) (intersect Coords, distance float64) {
  // Via https://stackoverflow.com/a/5205747

  // TODO: cleaner logic in implementation, needs to be several helper functions for testing

  // Query coords as floats for internal math
  xQuery := query.X
  yQuery := query.Y

  // Edge coordinate references
  x1 := e.Start.Pos.X
  y1 := e.Start.Pos.Y
  x2 := e.End.Pos.X
  y2 := e.End.Pos.Y

  // Slope of edge and perpendecular section
  mEdge := (y2 - y1) / (x2 - x1)
  mPerp := float64(-1) / mEdge

  //fmt.Print("<checkIntersect>")
  //fmt.Print(" query: {", xQuery, ",", yQuery, "}")
  //fmt.Print(", Edge{", x1, ",", y1, "}->{", x2, ",", y2, "}")
  //fmt.Print(", mEdge: ", mEdge, ", mPerp: ", mPerp)

  inPerp := false  // By default, not in perpendicular region
  isVert := math.IsInf(mPerp, 0)  // Is the perpendicular section vertical?

  // Determine if the query point lies in region perpendicular to the edge segment
  if (isVert) {  // Regions determined by x values
    if ((x1 < x2 && x1 < xQuery && xQuery < x2) || (x2 < x1 && xQuery < x1 && x2 < xQuery)) {
      inPerp = true
    }
  } else {  // Regions determined by y values
    // Relative y-coordinates of perpendicular lines at bounds of edge segment at x-value of query
    y1Rel := mPerp * (xQuery - x1) + y1
    y2Rel := mPerp * (xQuery - x2) + y2
    if ((y1Rel < y2Rel && y1Rel < yQuery && yQuery < y2Rel) || (y2Rel < y1Rel && yQuery < y1Rel && y2Rel < yQuery)) {
      inPerp = true
    }
  }

  //fmt.Print(", inPerp: ", inPerp, ", isVert: ", isVert)

  if (!inPerp) {  // Not in perpendicular segment or vertical; identify closest
    distance1 := query.Distance(e.Start.Pos)
    distance2 := query.Distance(e.End.Pos)
    //fmt.Println(", dist1: ", distance1, ", dist2: ", distance2)
    if (distance1 < distance2) {
      intersect = e.Start.Pos
      distance = distance1
    } else {
      intersect = e.End.Pos
      distance = distance2
    }
  } else {  // In perpendicular region; find intersection on Edge
    // Check for straight lines to ease calculations
    if (x1 == x2) {
      intersect.X = math.Round(x1)
      intersect.Y = math.Round(yQuery)

    } else if (y1 == y2) {
      intersect.X = math.Round(xQuery)
      intersect.Y = math.Round(y1)
    } else {
      intX := (mEdge * x1 - mPerp * xQuery + yQuery - y1) / (mEdge - mPerp)
      intY := mPerp * (intersect.X - xQuery) + yQuery
      intersect.X = math.Round(intX)
      intersect.Y = math.Round(intY)
    }
    distance = intersect.Distance(query)
  }
  //fmt.Print(", intersect: ", intersect, ", distance: ", distance)
  //fmt.Println()
  return
}

