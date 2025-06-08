package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)
	

var clients2 = make(map[*websocket.Conn]bool)
var clientsMutex2 = &sync.Mutex{}
var upgrader2 = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
func startServer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("error upgrading server from http to websocket", err)
		return
	}
	defer conn.Close()

	log.Println("New client has connected.")

	clientsMutex2.Lock()
	clients2[conn] = true
	clientsMutex2.Unlock()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error reading the message", err)
			break
		}
		if msgType == websocket.TextMessage {
			log.Printf("Received %s\n", msg)

			clientsMutex2.Lock()
			for client := range clients2 {
				if client != conn {
					err := client.WriteMessage(websocket.TextMessage, msg)
					if err != nil {
						log.Println("Write error: ", err)
						client.Close()
					}
				}
			}
			clientsMutex2.Unlock()
		}
	}
	clientsMutex2.Lock()
	delete(clients2, conn)
	clientsMutex2.Unlock()

}

func main() {
	http.HandleFunc("/ws", startServer)
	fmt.Println("Party started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting the server: ", err)
	}

}