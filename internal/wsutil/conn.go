package wsutil

import (
	"net"
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
	if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}
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