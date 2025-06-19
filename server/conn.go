package server

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeWait = 10 * time.Second

type peer struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func (p *peer) write(msgType int, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return p.ws.WriteMessage(msgType, data)
}