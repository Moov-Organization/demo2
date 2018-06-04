package sim2

import (
  "github.com/gorilla/websocket"
  "net/http"
	"log"

	"fmt"
)

// Message - struct to contain all data relevant to rendering a Car on the frontend.
type Message struct {
	Type            string `json:"type"`
	ID              string `json:"id"`
	X               string `json:"x"`
	Y               string `json:"y"`
	Orientation     string `json:"orientation"`
	State           string `json:"state"`
	Address         string `json:"address"`
	West            string `json:"west"`
	South           string `json:"south"`
	East            string `json:"east"`
	North           string `json:"north"`
}

// Message struct to handshake the connection type with the client
type HandshakeMessage struct {
	Testing       string `json:"testing"`
	MrmAddress    string `json:"mrmAddress"`
}

// ride struct to receive locations
type RideRequestMessage struct {
	From   string `json:"from"`
	To     string `json:"to"`
}


// WebSrv - container for web server variables for the simulator.
type WebSrv struct {
  // TODO: other fields here as necessary
  webChan chan Message  // Incoming car information from simulator
}

// TODO: find a way to move these into the WebSrv struct without violating handleConnection
var clients = make(map[*websocket.Conn]bool) // connected clients
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var ExistingMrmAddress string
var SendTestChain chan Ride
var Testing bool
// NewWebSrv - Constructor for a valid WebSrv object.
func NewWebSrv(web chan Message, existingMrmAddress string) *WebSrv {
  s := new(WebSrv)
  s.webChan = web
  ExistingMrmAddress = existingMrmAddress
  Testing = false
  return s
}

func NewTestChainWebSrv(web chan Message, sendTestChain chan Ride) *WebSrv {
	s := new(WebSrv)
	s.webChan = web
	Testing = true
	SendTestChain = sendTestChain
	return s
}

// LoopWebSrv - Begin the web server execution loop.
func (s *WebSrv) LoopWebSrv() {
  // Create a simple file server
  fs := http.FileServer(http.Dir("public"))
  http.Handle("/", fs)

  // Configure websocket route
  http.HandleFunc("/ws", handleConnections)

  // Start the server on localhost portNo and log any errors
  go func() {
		log.Printf("http server started on %s \n", ":8000")
		err := http.ListenAndServe(":8000", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
  }()

  // Handle any car info updates from World
  for {
    // Grab the next message from the broadcast channel
    msg := <-s.webChan
    //fmt.Println("Got msg:",msg)
    // Send it out to every client that is currently connected
    for client := range clients {
      err := client.WriteJSON(msg)
      if err != nil {
        log.Printf("error: %v", err)
        client.Close()
        delete(clients, client)
      }
    }
  }
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true
	if Testing {
		ws.WriteJSON(HandshakeMessage{Testing:"true"})
	} else {
		ws.WriteJSON(HandshakeMessage{Testing:"false", MrmAddress:ExistingMrmAddress})
	}
	for {
		var rideReqMsg RideRequestMessage
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&rideReqMsg)
		if Testing && rideReqMsg.To != "" && rideReqMsg.From != "" {
			fmt.Println("Received Ride Request", rideReqMsg.From, " ", rideReqMsg.To)
			SendTestChain <- Ride{rideReqMsg.From, rideReqMsg.To}
		}
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
	}
}
