package log

import (
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

// Logger is interface for output pod log
type Logger interface {
	io.WriteCloser
}

var upgrader = func() websocket.Upgrader {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	return upgrader
}()

// WsLogger output container log to websocket
type WsLogger struct {
	wsConn *websocket.Conn
}

// NewWsLogger create WsLogger
func NewWsLogger(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WsLogger, error) {
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	session := &WsLogger{
		wsConn: conn,
	}
	return session, nil
}

// Write wirte bytes
func (l *WsLogger) Write(p []byte) (n int, err error) {
	if err := l.wsConn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close ws connection
func (l *WsLogger) Close() error {
	return l.wsConn.Close()
}
