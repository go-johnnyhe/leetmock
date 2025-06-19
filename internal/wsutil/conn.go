package wsutil

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)


const waitTime = 10 * time.Second

type Peer struct {
	*websocket.Conn
	mu sync.Mutex
}

func NewPeer(conn *websocket.Conn) *Peer {
	return &Peer{
		Conn: conn,
	}
}

func (p *Peer) Write(msgType int, msg []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.SetWriteDeadline(time.Now().Add(waitTime))
	return p.WriteMessage(msgType, msg)
}