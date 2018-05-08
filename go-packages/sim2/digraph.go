package sim2

import (
	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
  "math"
  //"fmt"
)

// digraph - Describes an implementation of a simple weighted directed graph with underlying coords

// Vertex - struct for generic digraph vertex.
type Vertex struct {
  ID uint
  Pos Coords
  AdjEdges []*Edge
}

// Edge - struct for directed weighted edge in digraph.
type Edge struct {
  ID uint
  Start *Vertex
  End *Vertex
  Weight float64
}

// Digraph - struct for Digraph object.
type Digraph struct {
  Vertices map[uint]*Vertex  // map vertex ID to vertex reference
  Edges map[uint]*Edge  // map edge ID to edge reference
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
    line := strings.Split(scanner.Text()," ")

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
    x, y := splitCSV(line[1])
    vert.Pos.X = x
    vert.Pos.Y = y

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
      vert.AdjEdges = append(vert.AdjEdges, edge)
    }

    // Set starting edge weight based on distance
    for _, edge := range d.Edges {
      edge.Weight = edge.Start.Pos.Distance(edge.End.Pos)
    }
  }

  return d
}

// splitCSV - Split a CSV pair into constituent values.
func splitCSV(line string) (float64, float64) {
	numbers := strings.Split(line,",")
	if len(numbers) != 2 {
		log.Fatal("Line does not have format <number1>, <number2> ")
	}
	numberOne, err := strconv.Atoi(numbers[0])
	if err != nil {
		log.Fatal("Line does not have format <number1>, <number2> ")
	}
	numberTwo, err := strconv.Atoi(numbers[1])
	if err != nil {
		log.Fatal("Line does not have format <number1>, <number2> ")
	}
	return float64(numberOne), float64(numberTwo)
}

// ShortestPath - solve for the shortest deighted directional path from start to end vertex.
//   If not path can be found, return an empty slice and infinite distance.
//   Negative edge weights are not permitted.
func (g *Digraph) ShortestPath(startVertID, endVertID uint) (edgeIDs []uint, dist float64) {
  edgeIDs = make([]uint, 0)
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

    len := len(inverse)
    for idx := len - 1; idx >= 0; idx-- {
      edgeIDs = append(edgeIDs, inverse[idx])
    }
  }
  return
}
