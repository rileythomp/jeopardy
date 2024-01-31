package socket

import (
	"sync"

	"github.com/gorilla/websocket"
)

const (
	Info         = 4100
	Ok           = 4200
	BadRequest   = 4400
	Unauthorized = 4401
	ServerError  = 4500
)

type SafeConn struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func NewSafeConn(conn *websocket.Conn) *SafeConn {
	return &SafeConn{
		conn: conn,
	}
}

func (s *SafeConn) ReadMessage() (messageType int, p []byte, err error) {
	return s.conn.ReadMessage()
}

func (s *SafeConn) WriteJSON(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteJSON(v)
}

func (s *SafeConn) Close() error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}
