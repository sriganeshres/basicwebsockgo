package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Define a struct to represent a WebSocket client
type Client struct {
	conn *websocket.Conn
}

var (
	clients   = make(map[*Client]bool)
	broadcast = make(chan []byte)
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &Client{conn: conn}
	clients[client] = true

	// Print a message when someone connects
	fmt.Println("Client connected")

	// Infinite loop to read messages from the client
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			delete(clients, client)
			return
		}

		// Broadcast the message to all clients
		broadcast <- p

		// Print the received message to the console
		fmt.Printf("Received message: %s\n", p)
	}
}

func handleMessages() {
	for {
		// Get the next message from the broadcast channel
		message := <-broadcast

		// Send the message to all connected clients
		for client := range clients {
			err := client.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println(err)
				delete(clients, client)
				client.conn.Close()
			}
		}
	}
}

func main() {
	// Start a goroutine to handle incoming messages and broadcast them to all clients
	go handleMessages()

	// Handle WebSocket connections at the "/ws" endpoint
	http.HandleFunc("/ws", handleConnections)

	// Start the server on port 8080
	fmt.Println("WebSocket server running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err)
	}
}
