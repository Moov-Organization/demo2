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
)

// digraph - Describes an implementation of a simple weighted directed graph with underlying coords

// Vertex - struct for generic digraph vertex.
type Vertex struct {
  ID uint
  Pos Coords
  AdjEdges []Edge
  intersection *Intersection
  directionFromIntersection Direction
}

// Edge - struct for directed weighted edge in digraph.
type Edge struct {
  ID uint
  Start *Vertex
  End *Vertex
  Weight float64
}

// The number of directions at an waitingFor
const NumberOfDirections = 4

type Intersection struct {
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
      d.Edges[edge.ID] = edge
      vert.AdjEdges = append(vert.AdjEdges, *edge)
    }

    // Set starting edge weight based on distance
    for _, edge := range d.Edges {
      edge.Weight = edge.Start.Pos.Distance(edge.End.Pos)
    }
  }

	for scanner.Scan() {
		text := scanner.Text()
		numbers := splitLine(text," ", 4)
		var stopSign Intersection
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
func (g Digraph) shortestPath(startVertID, endVertID uint) (edges []Edge, dist float64) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	adjEdges := g.Vertices[startVertID].AdjEdges
	edge := adjEdges[uint(r1.Int() % len(adjEdges))]
	edges = append(edges, edge)

	adjEdges = edge.End.AdjEdges
	edge = adjEdges[uint(r1.Int() % len(adjEdges))]
	edges = append(edges, edge)

	adjEdges = edge.End.AdjEdges
	edge = adjEdges[uint(r1.Int() % len(adjEdges))]
	edges = append(edges, edge)

	adjEdges = edge.End.AdjEdges
	edge = adjEdges[uint(r1.Int() % len(adjEdges))]
	edges = append(edges, edge)

	adjEdges = edge.End.AdjEdges
	edge = adjEdges[uint(r1.Int() % len(adjEdges))]
	edges = append(edges, edge)

	// TODO: implement this in a concurrency-safe manner: do not modify underlying values!
  return
}

// closestEdgeAndCoord For coords within world space, find  closest coords on an edge on world graph
// Return coordinates of closest point on world graph, and corresponding edge ID in world struct
func (g Digraph) closestEdgeAndCoord(queryPoint Coords) (location Location) {
	// TODO: input sanitation/validation; error handling?
	// TODO: proper helper function breakdown of closestEdgeAndCoord

	shortestDistance := math.Inf(1)
	location.intersect = Coords{0, 0}

	// TODO: remove randomness caused by traversing equivalent closest edges with 'range' on map here
	for _, edge := range g.Edges {
		coord, dist := edge.checkIntersect(queryPoint)
		//fmt.Print("[", edgeIdx, "] <ClosestEdgeAndCoord>")
		//fmt.Print(" query: ", queryPoint, ", shortest: ", shortestDistance, ", new: ", coord, ", dist: ", dist)
		if dist < shortestDistance {
			shortestDistance = dist
			location.intersect = coord
			location.edge = *edge
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





