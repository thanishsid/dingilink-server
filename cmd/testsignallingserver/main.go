package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader to upgrade HTTP to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var clientsMutex = sync.Mutex{}

// Message defines the structure for signaling
type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func main() {
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Handle incoming WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clientsMutex.Lock()
	clients[ws] = true
	clientsMutex.Unlock()

	// Continuously listen for messages from the client
	for {
		var msg Message

		// Read the incoming message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading from websocket: %v", err)
			delete(clients, ws)
			break
		}

		// Broadcast the message to all other clients
		clientsMutex.Lock()
		for c := range clients {
			if c != ws {
				err := c.WriteJSON(msg)
				if err != nil {
					log.Printf("error writing to websocket: %v", err)
					c.Close()
					delete(clients, c)
				}
			}
		}
		clientsMutex.Unlock()
	}
}
