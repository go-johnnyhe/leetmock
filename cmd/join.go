/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var timer *time.Timer
var isWritingReceivedFile = false
var writeMutex sync.Mutex
var lastHash sync.Map

func fileHash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

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
		go readFile(conn)

		go monitorFile(conn)

		<-ctx.Done()
		fmt.Println("")
		fmt.Println("Goodbye!")
	},
}

func readFile(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading msg", err)
			return
		}
		parts := strings.SplitN(string(msg), "|", 2)
		if len(parts) != 2 {
			log.Printf("Received invalid message format: %s\n", string(msg))
			continue
		}

		filename:= parts[0]
		content := parts[1]

		writeMutex.Lock()
		isWritingReceivedFile = true
		if err = os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Printf("error writing this file: %s: %v\n", filename, err)
		} else {
			log.Printf("Received update to %s (%d bytes)\n", filename, len(content))
		}
		lastHash.Store(filename, fileHash([]byte(content)))

		time.Sleep(100 * time.Millisecond)
		isWritingReceivedFile = false
		writeMutex.Unlock()
	}
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
				log.Printf("FS-Event  %s  %s\n", event.Op.String(), event.Name)
				if !ok {
					return
				}
				if timer != nil {
					timer.Stop()
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) != 0 {
				filePath := event.Name
				base := filepath.Base(filePath)

				// Skip Vim swap/undo files only
				if strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".tmp") {
					return
				}

				// If it's a backup (ends with ~), sync the *real* file instead
				if strings.HasSuffix(base, "~") {
					filePath = strings.TrimSuffix(filePath, "~")
				}

				timer = time.AfterFunc(500*time.Millisecond, func() {
					sendFile(filePath, conn)
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
	if isWritingReceivedFile {
		fmt.Printf("Skipping send - currently writing received file\n")
		return
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("error reading the file: ", err)
		return
	}


	key := filepath.Base(filePath)
	newHash := fileHash(content)
	if prev, ok := lastHash.Load(key); ok && prev.(string) == newHash {
		return
	}
	lastHash.Store(key, newHash)
	message := fmt.Sprintf("%s|%s", key, content)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		log.Println("error writing the file: ", err)
	}
	fmt.Printf("Sent %s at: %s\n", filePath, time.Now().Format("15:04:05"))
}

func init() {
	rootCmd.AddCommand(joinCmd)
}
