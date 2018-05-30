package sim2

import (
  "os"
  "log"
  "bufio"
  "strings"
  "strconv"
  "math"
  "math/rand"
  "time"
  //"fmt"
)

// digraph - Describes an implementation of a simple weighted directed graph with underlying coords

// Vertex - struct for generic digraph vertex.
type Vertex struct {
  ID uint
  Pos Coords
  AdjEdges []*Edge
  intersection *Intersection
  directionFromIntersection Direction
}

// Edge - struct for directed weighted edge in digraph.
type Edge struct {
  ID uint
  Start *Vertex
  End *Vertex
  Weight float64
  Extends bool
}

// The number of directions at an waitingFor
const NumberOfDirections = 4

type Intersection struct {
  id uint
  entries [NumberOfDirections]EntryInfo
  intersectionType IntersectionType
}

// Entry Info - information about the entry to the waitingFor
type EntryInfo struct {
  present bool    // if there is an entry from the corresponding direction
  vertex *Vertex  // the vertex corresponding to the entry
}

type Direction int
const (
  West   Direction = 0
  South  Direction = 1
  East   Direction = 2
  North  Direction = 3
)

type IntersectionType int
const (
  NoIntersection  IntersectionType = 0
  StopSign        IntersectionType = 1
  StopLight       IntersectionType = 2
)

// Digraph - struct for Digraph object.
type Digraph struct {
  Vertices map[uint]*Vertex  // map vertex ID to vertex reference
  Edges map[uint]*Edge  // map edge ID to edge reference
  Intersections []*Intersection
}

// NewDigraph - Constructor for valid Digraph object.
func NewDigraph() *Digraph {
  d := new(Digraph)
  d.Vertices = make(map[uint]*Vertex)
  d.Edges = make(map[uint]*Edge)
  return d
}

// GetDigraphFromFile - Populate Digraph object from fomratted 'map' file.
func GetDigraphFromFile(fname string) (d *Digraph) {
  d = NewDigraph()

  file, err := os.Open(fname)
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)

  // Scan input file line by line
  for scanner.Scan() {
    text := scanner.Text()
    if text == "STOPSIGNS" {
      break
    }
    line := strings.Split(text," ")

    // Fetch id of Vertex described on this line
    idRead, _ := strconv.Atoi(line[0])
    id := uint(idRead)

    // Construct and populate new Vertex
    if _, ok := d.Vertices[id]; !ok {
      d.Vertices[id] = new(Vertex)
    }
    vert := d.Vertices[id]
    vert.ID = id
    // Parse input Vertex coordinates
    numbers := splitLine(line[1], ",",2)
    vert.Pos.X = numbers[0]
    vert.Pos.Y = numbers[1]

    // Parse adjacent vertices
    for _, point := range line[2:] {
      // Construct and populate new adjacent Vertex
      edgeExtends := false
      if point[len(point)-1:] == "e"{
      	edgeExtends = true
				point = point[:len(point)-1]
			}
      idReadNext, _ := strconv.Atoi(point)
      idNext := uint(idReadNext)
      if _, ok := d.Vertices[idNext]; !ok {
        d.Vertices[idNext] = new(Vertex)
      }
      vertNext := d.Vertices[idNext]

      // Construct and populate new edge between adjacent vertices
      edge := new(Edge)
      edge.ID = uint(len(d.Edges))
      edge.Start = vert
      edge.End = vertNext
      edge.Extends = edgeExtends
      d.Edges[edge.ID] = edge
      vert.AdjEdges = append(vert.AdjEdges, edge)
    }

    // Set starting edge weight based on distance
    for _, edge := range d.Edges {
      edge.Weight = edge.Start.Pos.Distance(edge.End.Pos)
    }
  }

  intersectionCount := 0
  for scanner.Scan() {
    text := scanner.Text()
    if text == "STOPLIGHTS" {
      break
    }
    numbers := splitLine(text," ", 4)
    var stopSign Intersection
    stopSign.id = uint(intersectionCount)
    stopSign.intersectionType = StopSign
    for direction, number := range numbers {
      if number >= 0 {
        stopSign.entries[direction].present = true
        stopSign.entries[direction].vertex = d.Vertices[uint(number)]
        d.Vertices[uint(number)].intersection = &stopSign
        d.Vertices[uint(number)].directionFromIntersection = Direction(direction)
      }
    }
    d.Intersections = append(d.Intersections, &stopSign)
    intersectionCount++
  }

  for scanner.Scan() {
    text := scanner.Text()
    numbers := splitLine(text," ", 4)
    var stopLight Intersection
    stopLight.id = uint(intersectionCount)
    stopLight.intersectionType = StopLight
    for direction, number := range numbers {
      if number >= 0 {
        stopLight.entries[direction].present = true
        stopLight.entries[direction].vertex = d.Vertices[uint(number)]
        d.Vertices[uint(number)].intersection = &stopLight
        d.Vertices[uint(number)].directionFromIntersection = Direction(direction)
      }
    }
    d.Intersections = append(d.Intersections, &stopLight)
    intersectionCount++
  }
  return d
}

// splitCSV - Split a CSV pair into constituent values.
func splitLine(line string, separator string, length int) (numbers []float64) {
  numbersInString := strings.Split(line, separator)
  if len(numbersInString) != length {
    log.Fatalf("Line does not have %d numbers, it has %d numbers", length, len(numbers))
  }

  for _, numberInString := range numbersInString {
    number, err := strconv.Atoi(numberInString)
    if err != nil {
      log.Fatalf("%s in %s is not a number", numberInString, numbersInString)
    }
    numbers = append(numbers, float64(number))
  }
  return
}

// ShortestPath - solve for the shortest deighted directional path from start to end vertex.
//   If not path can be found, return an empty slice and infinite distance.
//   Negative edge weights are not permitted.
func (g *Digraph) shortestPath(startVertID, endVertID uint) (edges []Edge, dist float64) {
  edgeIDs := make([]uint, 0)
  dist = math.Inf(1)

  // NOTE: this shortest path implementation follows ShortestPath's algorithm

  // Receiver validity check
  if g.Vertices == nil || g.Edges == nil {
    log.Println("err: ShortestPath - digraph not initialized")
    return
  }

  // Bounds check; ensure start and end vertices are in the graph
  _, ok0 := g.Vertices[startVertID]
  _, ok1 := g.Vertices[endVertID]
  if !ok0 || !ok1 {
    log.Println("err: ShortestPath - invalid start or end vertex")
    return
  }

  // Mark all vertices as 'unvisited'
  unvisited := make(map[uint]bool)  // Set of vertex IDs
  for idx := range g.Vertices {
    unvisited[idx] = true
  }
  //fmt.Println("unvisited:", unvisited)

  // Define local struct type to map minimum distance to previous vertex
  type minDist struct {
    dist float64  // Minimum distance to start from this vertex through previous vertex
    prev uint  // ID of previous vertex for this distance
  }

  // Assign each vertex a tentative distance value from start: +Inf by default
  distances := make(map[uint]minDist)
  for idx := range g.Vertices {
    distances[idx] = minDist{math.Inf(1), 0}
  }
  distances[startVertID] = minDist{float64(0), 0}  // starting vertex has no distance to itself

  // Iterate over all vertices, starting at the query vertex, until end vertex is found
  currID := startVertID
  for ; len(unvisited) != 0 ; {
    // Create a local record of the shortest distance in this iteration (optimization)
    shortest := minDist{math.Inf(1), 0}

    // Consider all unvisited neighbors of the current vertex
    for _, adjEdge := range g.Vertices[currID].AdjEdges {
      neighborID := adjEdge.End.ID  // Identify neighbor vertex

      // Only consider if unvisited
      if _, ok := unvisited[neighborID]; ok {

        // Determine distance to start vertex
        localdist := distances[currID].dist + adjEdge.Weight
        //fmt.Println("currID", currID, "neighborID", neighborID, "adjEdge", adjEdge)
        //fmt.Println("localdist", localdist, "currdist", distances[currID].dist, "weight", adjEdge.Weight)

        // If this distance is shorter than recorded distance, update
        if localdist < distances[neighborID].dist {
          distances[neighborID] = minDist{localdist, currID}
        }

        // If this is the shortest distance for the current vertex, note so
        if localdist < shortest.dist {
          shortest = minDist{localdist, neighborID}
        }
      }
    }

    // If identified shortest is valid, use it as next vertex to visit
    var nextID uint
    if !math.IsInf(shortest.dist, 1) {
      nextID = uint(shortest.prev)
    } else {  // Otherwise, search graph for next closest vertex
      nextClosest := minDist{math.Inf(1), 0}  // Identify next closest vertex

      // Walk the unvisited vertices to find the next closest vertex
      for idx := range unvisited {
        if distances[idx].dist < nextClosest.dist {
          nextClosest = minDist{distances[idx].dist, idx}
        }
      }

      // Assign next vertex as the closest unvisited vertex
      nextID = uint(nextClosest.prev)
    }

    // Identify current vertex as 'visited' and move to next vertex
    delete(unvisited, currID)
    currID = nextID
  }

  //fmt.Println("DONE")

  // Determine if a valid path was found
  if !math.IsInf(distances[endVertID].dist, 1) {
    inverse := make([]uint, 0)
    dist = float64(0)

    // Build the path backwards from the destination vertex
    inverse = append(inverse, endVertID)
    for curr := endVertID; curr != startVertID; curr = distances[curr].prev {
      inverse = append(inverse, distances[curr].prev)
      dist += distances[curr].dist
      //fmt.Println("cd", distances[curr].dist)
    }

    // Invert the path back for the path from start

    length := len(inverse)
    for idx := length - 1; idx >= 0; idx-- {
      edgeIDs = append(edgeIDs, inverse[idx])
    }
  }

  for i :=0 ; i < len(edgeIDs)-1; i++ {
    vertex := g.Vertices[edgeIDs[i]]
    for _, adjEdge := range vertex.AdjEdges {
      if edgeIDs[i+1] == adjEdge.End.ID {
        edges = append(edges, *adjEdge)
        break
      }
    }
  }

  return
}

// closestEdgeAndCoord For coords within world space, find  closest coords on an edge on world graph
// Return coordinates of closest point on world graph, and corresponding edge ID in world struct
func (g Digraph) closestEdgeAndCoord(queryPoint Coords) (location Location) {
  // TODO: input sanitation/validation; error handling?
  // TODO: proper helper function breakdown of closestEdgeAndCoord

  shortestDistance := math.Inf(1)
  location.intersect = Coords{0, 0}

  //fmt.Println("<ClosestEdgeAndCoord>")
  //fmt.Println(" query: ", queryPoint)

  // TODO: remove randomness caused by traversing equivalent closest edges with 'range' on map here
  for _, edge := range g.Edges {
    coord, dist := edge.checkIntersect(queryPoint)
    //fmt.Print("[", edge.Start.ID, ", ", edge.End.ID, "]: ")
    //fmt.Print("shortest: ", location.intersect, "@", shortestDistance, ", new: ", coord, "@", dist)
    if dist < shortestDistance {
      shortestDistance = dist
      location.intersect = coord
      location.edge = *edge
      //fmt.Print(" (new shortest: ", coord, "@", dist, ")")
    }
    //fmt.Println()
  }
  //fmt.Print("Shortest from query ", queryPoint, ": ", location.intersect, " on {")
  //fmt.Println(location.edge.Start.ID, ",", location.edge.End.ID, "}");
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

  //fmt.Println("<checkIntersect>")
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
  } else {  // In perpendicular region; find waitingFor on Edge
    // Check for straight lines to ease calculations
    if (x1 == x2) {
      intersect.X = math.Round(x1)
      intersect.Y = math.Round(yQuery)

    } else if (y1 == y2) {
      intersect.X = math.Round(xQuery)
      intersect.Y = math.Round(y1)
    } else {
      intX := (mEdge * x1 - mPerp * xQuery + yQuery - y1) / (mEdge - mPerp)
      intY := mPerp * (intX - xQuery) + yQuery
      intersect.X = math.Round(intX)
      intersect.Y = math.Round(intY)
    }
    distance = intersect.Distance(query)
  }
  //fmt.Print(", intersect: ", intersect, ", distance: ", distance)
  //fmt.Println()
  return
}

func (e *Edge) unitVector() (c Coords){
  c = e.Start.Pos.UnitVector(e.End.Pos)
  return
}

func (g Digraph) getRandomEdge() (edge Edge) {
  s1 := rand.NewSource(time.Now().UnixNano())
  r1 := rand.New(s1)
  edge = *g.Edges[uint(r1.Int() % len(g.Edges))]
  return
}
