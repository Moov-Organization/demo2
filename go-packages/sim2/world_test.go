package sim2



/*   World stubs    */
type MockWorld struct {
	getRandomEdgeStruct GetRandomEdgeStruct
	closestEdgeAndCoordStruct ClosestEdgeAndCoordStruct
	shortestpathStruct ShortestPathStruct
	getVertexStruct GetVertexStruct
}

type GetRandomEdgeStruct struct {
	returnEdge Edge
	function func() (Edge)
	calls uint
}

func (mockWorldAPI *MockWorld) getRandomEdge() (edge Edge) {
	mockWorldAPI.getRandomEdgeStruct.calls++
	if mockWorldAPI.getRandomEdgeStruct.function != nil {
		return mockWorldAPI.getRandomEdgeStruct.function()
	}
	edge = mockWorldAPI.getRandomEdgeStruct.returnEdge
	return
}

type GetVertexStruct struct {
	paramId uint
	returnVertex Vertex
	function func() (Vertex)
	calls uint
}

func (mockWorldAPI *MockWorld) getVertex(id uint) (returnVertex Vertex) {
	mockWorldAPI.getVertexStruct.calls++
	mockWorldAPI.getVertexStruct.paramId = id
	if mockWorldAPI.getVertexStruct.function != nil {
		return mockWorldAPI.getVertexStruct.function()
	}
	returnVertex = mockWorldAPI.getVertexStruct.returnVertex
	return
}

type ClosestEdgeAndCoordStruct struct {
	paramQueryPoint Coords
	returnLocation Location
	function func(Coords) (Location)
	calls uint
}

func (mockWorldAPI *MockWorld) closestEdgeAndCoord(queryPoint Coords) (location Location) {
	mockWorldAPI.closestEdgeAndCoordStruct.calls++
	mockWorldAPI.closestEdgeAndCoordStruct.paramQueryPoint = queryPoint
	if mockWorldAPI.closestEdgeAndCoordStruct.function != nil {
		return mockWorldAPI.closestEdgeAndCoordStruct.function(queryPoint)
	}
	location = mockWorldAPI.closestEdgeAndCoordStruct.returnLocation
	return
}

type ShortestPathStruct struct {
	paramStartVertID uint
	paramEndVertID uint
	returnEdges []Edge
	returnDistance float64
	function func(uint, uint) ([]Edge, float64)
	calls uint
}

func (mockWorldAPI *MockWorld) shortestPath(startVertID, endVertID uint) (edges []Edge, distance float64) {
	mockWorldAPI.shortestpathStruct.calls++
	mockWorldAPI.shortestpathStruct.paramStartVertID = startVertID
	mockWorldAPI.shortestpathStruct.paramEndVertID = endVertID
	if mockWorldAPI.shortestpathStruct.function != nil {
		return mockWorldAPI.shortestpathStruct.function(startVertID, endVertID)
	}
	edges = mockWorldAPI.shortestpathStruct.returnEdges
	distance = mockWorldAPI.shortestpathStruct.returnDistance
	return
}
