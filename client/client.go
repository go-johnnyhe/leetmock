package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

func startClient() {

	// step 1: read user input from terminal
	conn, _, err := websocket.DefaultDialer.Dial("wss://6b33-63-208-141-34.ngrok-free.app/ws", nil)
	if err != nil {
		fmt.Println("Dial error: ", err)
		return
	}
	defer conn.Close()
	go func() {
		for {
			fmt.Println("Please enter your message: ")
			reader := bufio.NewReader(os.Stdin)
			msg, _ := reader.ReadString('\n')
			conn.WriteMessage(websocket.TextMessage, []byte(msg))
			// fmt.Println(msg)
		}
	}()

	// step 2: receive message from websockets
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read error: ", err)
				return
			}
			fmt.Printf("Received: %s", message)
		}
	}()

	select {}
}

func main() {
	startClient()
}
