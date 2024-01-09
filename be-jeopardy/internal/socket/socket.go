package socket

import (
	"sync"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/jeopardy"
)

type SafeConn struct {
	mu   sync.Mutex
	Conn jeopardy.SafeConn
}

func NewSafeConn(conn jeopardy.SafeConn) *SafeConn {
	return &SafeConn{
		Conn: conn,
	}
}

func (s *SafeConn) ReadMessage() (messageType int, p []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Conn.ReadMessage()
}

func (s *SafeConn) WriteJSON(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Conn.WriteJSON(v)
}

func (s *SafeConn) Close() error {
	if s.Conn == nil {
		return nil
	}
	return s.Conn.Close()
}
