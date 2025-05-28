package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var clientsMutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	//upgrade HTTP connection to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	defer conn.Close()

	//add new client to map
	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	log.Println("New client connected.")

	//keep listening for messages
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		if messageType == websocket.TextMessage {
			log.Printf("Received: %s\\n", p)

			clientsMutex.Lock()
			for client := range clients {
				if client != conn {
					err := client.WriteMessage(websocket.TextMessage, p)
					if err != nil {
						log.Println("Write error: ", err)
						client.Close()
					}
				}
			}
			clientsMutex.Unlock()
		}

		// if err := conn.WriteMessage(messageType, p); err != nil {
		// 	log.Println("Write error:", err)
		// 	return
		// }
	}

	clientsMutex.Lock()
	delete(clients, conn)
	clientsMutex.Unlock()
	log.Println("Client disconnected")
}

func main() {
	http.HandleFunc("/ws", handler)
	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	// client()
}
