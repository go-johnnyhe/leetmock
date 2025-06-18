package client

import (
    "crypto/sha256"
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

func (c *Client) Start() {
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
	message := fmt.Sprintf("%s|%s", key, content)

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
			log.Println("error reading msg: ", err)
			return
		}

		parts := strings.SplitN(string(msg), "|", 2)
		if len(parts) != 2 {
			log.Printf("Received invalid message format: %s\n", string(msg))
			continue
		}

		filename := filepath.Base(parts[0])
		if strings.Contains(filename, "..") || strings.HasPrefix(filename, "/") {
			log.Printf("invalid name: %s\n", filename)
			continue
		}
		content := parts[1]

		c.writeMutex.Lock()
		c.isWritingReceivedFile = true

		if err = os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Printf("error writing this file: %s: %v\n", filename, err)
		} else{
			log.Printf("received updates to %s\n", filename)
		}

		c.lastHash.Store(filename, fileHash([]byte(content)))
		time.Sleep(100*time.Millisecond)
		c.isWritingReceivedFile = false
		c.writeMutex.Unlock()
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
	
	select{}

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