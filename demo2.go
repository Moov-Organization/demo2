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
  w := sim2.GetWorldFromFile("maps/4by4.map")
  w.Fps = float64(60)

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

  // Instantiate cars
  numCars := uint(1)
  cars := make([]*sim2.Car, numCars)
  for i := uint(0); i < numCars; i++ {
    // Request to register new car from World
    syncChan, updateChan, ok := w.RegisterCar(i)
    if !ok {
      log.Printf("error: failed to register car")
    }

    scanner.Scan()
    scanner.Scan()
    carPrivateKey := scanner.Text()
    // Construct car
    cars[i] = sim2.NewCar(i, w, syncChan, updateChan, existingMrmAddress, carPrivateKey)
  }

  // Instantiate JSON web output
  webChan, ok := w.RegisterWeb()
  if !ok {
    log.Printf("error: failed to register web output")
  }
  web := sim2.NewWebSrv(webChan)

  // Begin World operation
  go w.LoopWorld()

  // Begin Car operation
  for i := uint(0); i < numCars; i++ {
    go cars[i].CarLoop()
  }

  // Begin JSON web output operation
  go web.LoopWebSrv()

  select{}  // Do work in the coroutines, main has nothing left to do
}
//*/
