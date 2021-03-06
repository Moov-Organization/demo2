package main

import (
  "demo2/go-packages/sim2"
  "log"
  "fmt"
  "os"
  "bufio"
  "flag"
)

// TODO: remove commented-out test prints and make proper test files

func main() {
  fmt.Println("Starting demo2 simulation")
  testingFlagPtr := flag.Bool("testing", true, "a boolean to turn on testing")
  portFlagPtr := flag.String("port", "8000", "a string to hold port number")
  flag.Parse()

  // Instantiate world
  graph := sim2.GetDigraphFromFile("maps/final.map")
  world := sim2.NewWorld(25, graph)

  // Instantiate JSON web output
  webChan, ok := world.RegisterWeb()
  if !ok {
    log.Fatalln("error: failed to register web output")
  }

  var web *sim2.WebSrv
  var existingMrmAddress string
  var scanner *bufio.Scanner
  var testChain *sim2.TestChain
  if (!*testingFlagPtr) {
    keys, err := os.Open("keys.txt")
    if err != nil {
      log.Fatal(err)
    }
    defer keys.Close()

    scanner = bufio.NewScanner(keys)

    scanner.Scan()
    scanner.Scan()
    //address of the deployed ferris contract
    existingMrmAddress = scanner.Text()

    web = sim2.NewWebSrv(webChan, existingMrmAddress)
  } else {
   testChain = sim2.NewTestChain()
    web = sim2.NewTestChainWebSrv(webChan, testChain.RecvServer)
  }
  // Instantiate cars
  numCars := uint(6)
  cars := make([]*sim2.Car, numCars)
  for i := uint(0); i < numCars; i++ {
    // Request to register new car from World
    id, syncChan, updateChan, ok := world.RegisterCar()
    if !ok {
      log.Fatalln("error: failed to register car")
    }
		graph = sim2.GetDigraphFromFile("maps/final.map")
    if (!*testingFlagPtr) {
      scanner.Scan()
      scanner.Scan()
      carPrivateKey := scanner.Text()
      eth := sim2.NewEthApi(existingMrmAddress, carPrivateKey)
      cars[i] = sim2.NewCar(id, graph, eth, syncChan, updateChan, webChan)
    } else {
    	fmt.Println("TESTING")
      testchainApi := testChain.RegisterBlockchainInteractor()
      cars[i] = sim2.NewCar(id, graph, testchainApi, syncChan, updateChan, webChan)
    }
  }
	if (*testingFlagPtr) {
		testChain.StartTestChain()
	}
  // Begin World operation
  go world.LoopWorld()

  // Begin Car operation
  for i := uint(0); i < numCars; i++ {
    go cars[i].CarLoop()
  }

  // Begin JSON web output operation
  go web.LoopWebSrv(*portFlagPtr)

  select{}  // Do work in the coroutines, main has nothing left to do
}
//*/
