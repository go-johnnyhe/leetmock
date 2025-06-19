package client

import (
	"context"
    "crypto/sha256"
	"encoding/base64"
    "encoding/hex"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    
    "github.com/fsnotify/fsnotify"
    "github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	timer *time.Timer
	timerMutex sync.Mutex
	isWritingReceivedFile bool
	writeMutex sync.Mutex
	lastHash sync.Map
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client {
		conn: conn,
	}
}

func (c *Client) Start(ctx context.Context) {
	go c.readLoop()
	go c.monitorFiles()
}

func fileHash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (c *Client) SendFile(filePath string) {
	c.writeMutex.Lock()
	isWriting := c.isWritingReceivedFile
	c.writeMutex.Unlock()

	if isWriting {
		log.Println("skipping send - currently writing a received file")
		return
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}
	if fileInfo.Size() > 10 * 1024 * 1024 {
		log.Printf("File %s too large (%d bytes)", filePath, fileInfo.Size())
		return
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("error reading the file: ", err)
		return
	}

	key := filepath.Base(filePath)
	newHash := fileHash(content)

	if prevContent, ok := c.lastHash.Load(key); ok && prevContent.(string) == newHash {
		return
	}

	c.lastHash.Store(key, newHash)
	encodedContent := base64.StdEncoding.EncodeToString(content)
	message := fmt.Sprintf("%s|%s", key, encodedContent)

	if err := c.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		log.Println("error writing the file: ", err)
		return
	}

	fmt.Printf("Sent %s at: %s\n", filePath, time.Now().Format("15:04:05"))
}

func (c *Client) readLoop() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection lost: %v", err)
			}
			return
		}
		if len(msg) > 10 * 1024 * 1024 {
			log.Printf("message too large: %d bytes", len(msg))
			continue
		}

		parts := strings.SplitN(string(msg), "|", 2)
		if len(parts) != 2 {
			log.Printf("Received invalid message format: %s\n", string(msg))
			continue
		}

		// check if file path is clean
		filename := filepath.Base(parts[0])
		cleanPath := filepath.Clean(filename)
		if cleanPath != filename || strings.Contains(filename, "..") || strings.HasPrefix(filename, "/") {
			log.Printf("invalid name: %s\n", filename)
			continue
		}
		decodedContent, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			log.Printf("error decoding content for %s: %v\n", filename, err)
			continue
		}

		c.writeMutex.Lock()
		c.isWritingReceivedFile = true
		c.writeMutex.Unlock()

		func() {
			defer func() {
				c.writeMutex.Lock()
				c.isWritingReceivedFile = false
				c.writeMutex.Unlock()
			}()
			if err = os.WriteFile(filename, decodedContent, 0644); err != nil {
				log.Printf("error writing this file: %s: %v\n", filename, err)
			} else{
				log.Printf("received updates to %s\n", filename)
			}
		}()
		c.lastHash.Store(filename, fileHash(decodedContent))
	}
}

func (c *Client) monitorFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("failed to create a watcher: ", err)
		return
	}
	defer watcher.Close()

	go c.processFileEvents(watcher)
	
	watchPath := "./"
	if err := watcher.Add(watchPath); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Watching for changes under the directory: ", watchPath)
	
	// select{}

}

func (c *Client) processFileEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <- watcher.Events:
			if !ok {
				return
			}
			log.Printf("FS-Event %s %s\n", event.Op.String(), event.Name)

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) != 0 {
				c.handleFileEvent(event)
			}
		case err, ok := <- watcher.Errors:
			if !ok {
				return
			}
			log.Println("error with watcher:", err)
		}
	}
}

func (c *Client) handleFileEvent(event fsnotify.Event) {
	filePath := event.Name
	base := filepath.Base(filePath)

	// skip vim swap/undo files
	if strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".tmp") {
		return
	}

	if strings.HasSuffix(base, "~") {
		filePath = strings.TrimSuffix(filePath, "~")
	}

	c.timerMutex.Lock()
	if c.timer != nil {
		c.timer.Stop()
	}

	c.timer = time.AfterFunc(500*time.Millisecond, func() {
		c.SendFile(filePath)
	})
	c.timerMutex.Unlock()
}