package sim2


/*   Eth API stubs    */
type MockEthAPI struct {
	getAddressStruct GetAddressStruct
	acceptRequestStruct AcceptRequestStruct
	getLocationStruct GetLocationStruct
}

type GetAddressStruct struct{
	returnAvailable bool
	returnAddress string
	function func() (bool, string)
	calls uint
}

func (mockEthApi *MockEthAPI) GetRideAddressIfAvailable() (available bool, address string) {
	mockEthApi.getAddressStruct.calls++
	if mockEthApi.getAddressStruct.function != nil {
		return mockEthApi.getAddressStruct.function()
	}
	available = mockEthApi.getAddressStruct.returnAvailable
	address = mockEthApi.getAddressStruct.returnAddress
	return
}

type AcceptRequestStruct struct {
	paramAddress string
	returnStatus bool
	function func(string) (bool)
	calls uint
}

func (mockEthApi *MockEthAPI) AcceptRequest(address string) (status bool) {
	mockEthApi.acceptRequestStruct.calls++
	mockEthApi.acceptRequestStruct.paramAddress = address
	if mockEthApi.acceptRequestStruct.function != nil {
		return mockEthApi.acceptRequestStruct.function(address)
	}
	status = mockEthApi.acceptRequestStruct.returnStatus
	return
}

type GetLocationStruct struct {
	paramAddress string
	returnFrom string
	returnTo string
	function func(string) (string, string)
	calls uint
}

func (mockEthApi *MockEthAPI) GetLocations(address string) (from string, to string) {
	mockEthApi.getLocationStruct.calls++
	mockEthApi.getLocationStruct.paramAddress = address
	if mockEthApi.getLocationStruct.function != nil {
		return mockEthApi.getLocationStruct.function(address)
	}
	from = mockEthApi.getLocationStruct.returnFrom
	to = mockEthApi.getLocationStruct.returnTo
	return
}