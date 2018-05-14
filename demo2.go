package main

import (
  "demo2/go-packages/sim2"
  "log"
  "fmt"
  "os"
  "bufio"
)

// TODO: remove commented-out test prints and make proper test files

func main() {
  fmt.Println("Starting demo2 simulation")

  // Instantiate world
  world := sim2.GetWorldFromFile("maps/4by4.map")
  world.Fps = float64(50)

  keys, err := os.Open("keys.txt")
  if err != nil {
    log.Fatal(err)
  }
  defer keys.Close()

  scanner := bufio.NewScanner(keys)

  scanner.Scan()
  scanner.Scan()
  //address of the deployed ferris contract
  existingMrmAddress := scanner.Text()

/*
  g := sim2.GetDigraphFromFile("maps/4by4.map")
  a, b := g.ShortestPath(7, 14)
  fmt.Println("path", a, "dist", b)
*/

  // Instantiate cars
  numCars := uint(2)
  cars := make([]*sim2.Car, numCars)
  for i := uint(0); i < numCars; i++ {
    // Request to register new car from World
    syncChan, updateChan, ok := world.RegisterCar(i)
    if !ok {
      log.Fatalln("error: failed to register car")
    }

    scanner.Scan()
    scanner.Scan()
    carPrivateKey := scanner.Text()
    // Construct car
    eth := sim2.NewEthApi(existingMrmAddress, carPrivateKey)
    cars[i] = sim2.NewCar(i, world, eth, syncChan, updateChan)
  }

  // Instantiate JSON web output
  webChan, ok := world.RegisterWeb()
  if !ok {
    log.Fatalln("error: failed to register web output")
  }
  web := sim2.NewWebSrv(webChan, existingMrmAddress)

  // Begin World operation
  go world.LoopWorld()

  // Begin Car operation
  for i := uint(0); i < numCars; i++ {
    go cars[i].CarLoop()
  }

  // Begin JSON web output operation
  go web.LoopWebSrv()

  select{}  // Do work in the coroutines, main has nothing left to do
}
