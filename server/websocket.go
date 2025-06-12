package server

import (
	"net/http"
	"sync"
	"fmt"
	"log"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var clientsMutex = &sync.Mutex{}
var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartServer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to websocket connection: ", err)
		return
	}

	log.Println("Connected to websocket!")

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	defer func() {
		conn.Close()
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		log.Printf("Client disconnected. Total clients now: %d", len(clients))
	}()


	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message from the websocket: ", err)
			break
		}
		if msgType == websocket.TextMessage {
			log.Printf("Message: %s", string(msg))

			clientsMutex.Lock()
			for client := range clients {
				if client != conn {
					err := client.WriteMessage(msgType, msg)
					if err != nil {
						fmt.Println("Error writing message to other clients: ", err)
					}
				}
			}
			clientsMutex.Unlock()
		}
	}
}