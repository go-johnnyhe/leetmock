/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var timer *time.Timer

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join <session-url>",
	Short: "Join an existing collaborative coding session",
	Long: `Join a collaborative coding session by connecting to the provided URL.

This will:
- Connect to the session via WebSocket
- Sync shared files to ./shared/ directory
- Enable real-time file synchronization

Example:
  leetmock join https://abc123.trycloudflare.com

The session URL comes from whoever ran 'leetmock start'.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("join called")
		if len(args) != 1 {
			fmt.Println("Error: this takes exactly one url")
			cmd.Usage()
			return
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		sessionUrl := args[0]
		wsURL := strings.Replace(sessionUrl, "https://", "wss://", 1)
		if !strings.HasSuffix(wsURL, "/ws") {
			wsURL = strings.TrimSuffix(wsURL, "/") + "/ws"
		}
		fmt.Println("Starting your mock interview session ...")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			fmt.Println("Error making connection", err)
			return
		}
		defer conn.Close()
		// Read channel
		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					fmt.Println("Error reading msg", err)
					return
				}
				fmt.Printf("Received: %s\n", msg)
			}
		}()

		go monitorFile(conn)
		// Write channel
		// go func() {
		// 	for {
		// 		fmt.Println("Please type a message in:")
		// 		reader := bufio.NewReader(os.Stdin)
		// 		input, _ := reader.ReadString('\n')
		// 		msg := strings.TrimSpace(input)
		// 		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		// 			fmt.Println("Error writing message", err)
		// 		}
		// 	}
		// }()

		<-ctx.Done()
		fmt.Println("")
		fmt.Println("Goodbye!")
	},
}


func monitorFile(conn *websocket.Conn) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("error creating a watcher: ", err)
		return
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <- watcher.Events:
				if !ok {
					return
				}
				if timer != nil {
					timer.Stop()
				}
				if event.Has(fsnotify.Write) {
					filepath := event.Name
					if strings.HasSuffix(filepath, ".tmp") || strings.HasSuffix(filepath, ".swp") {
						fmt.Println("ignoring temp file: ", filepath)
						continue
					}
					timer = time.AfterFunc(500*time.Millisecond, func() {
						sendFile(filepath, conn)
					})
				}
			case err, ok := <- watcher.Errors:
				if !ok {
					return
				}
				log.Println("error with the watcher: ", err)
			}
		}
	}()

	watchPath := "./"
	if err := watcher.Add(watchPath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching for changes under the directory: ", watchPath)
}

func sendFile(filePath string, conn *websocket.Conn) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("error reading the file: ", err)
		return
	}
	fmt.Println(string(content))

	if err := conn.WriteMessage(websocket.TextMessage, content); err != nil {
		log.Println("error writing the file: ", err)
	}
	fmt.Println("Sent at: ", time.Now())
}

func init() {
	rootCmd.AddCommand(joinCmd)
}
