package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

// make websocket connection
// have simultaneously read and write

func MakeConnection() {
	conn, _, err := websocket.DefaultDialer.Dial("wss://6b33-63-208-141-34.ngrok-free.app/ws", nil)
	if err != nil {
		fmt.Println("Error making connection: ", err)
		return
	}
	defer conn.Close()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading the message: ", err)
				return
			}
			fmt.Println("Received:", string(msg))
		}
	}()

	go func() {
		for {
			fmt.Println("type shit in:")
			reader := bufio.NewReader(os.Stdin)

			input, _ := reader.ReadString('\n')
			msg := strings.TrimSpace(input)

			conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}()

	select{}
}