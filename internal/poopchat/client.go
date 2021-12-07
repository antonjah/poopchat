package poopchat

import (
	"context"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	maxBufferSize  = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxBufferSize,
	WriteBufferSize: maxBufferSize,
}

// Client is a middleman between the websocket connection and the server.
type Client struct {
	sessionID string
	server    *Server
	conn      *websocket.Conn
	send      chan []byte
}

func (c *Client) WriteWelcomeMessage() error {
	return c.conn.WriteMessage(websocket.TextMessage, []byte("Welcome to poopchat!"))
}

func (c *Client) Read(ctx context.Context) {
	logger := ctx.Value(SessionLoggerKey).(*log.Entry)

	defer func() {
		c.server.unregister <- c
		c.conn.Close()
		logger.Info("Session closed")
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(
		func(string) error {
			return c.conn.SetReadDeadline(time.Now().Add(pongWait))
		})
	for {
		messageType, message, err := c.conn.ReadMessage()
		switch messageType {
		case websocket.TextMessage:
			logger.Info("Received text message")
		}

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.WithError(err).Error("Socket read error")
			}
			break
		}
		c.server.broadcast <- message
	}
}

func (c *Client) Write(ctx context.Context) {
	logger := ctx.Value(SessionLoggerKey).(*log.Entry)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err = w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.WithError(err).Error("Failed to set deadline")
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.WithError(err).Error("Failed to write ping message")
				return
			}
		}
	}
}

func Serve(server *Server, w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value(SessionLoggerKey).(*log.Entry)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade connection to websocket")
		return
	}

	sessionID, ok := r.Context().Value(SessionIDKey).(string)
	if !ok {
		logger.Error("Failed to read session ID")
	}

	client := &Client{sessionID: sessionID, server: server, conn: conn, send: make(chan []byte, 256)}
	client.server.register <- client

	logger.Info("Session registered")

	if err = client.WriteWelcomeMessage(); err != nil {
		logger.WithError(err).Error("Failed to post welcome message")
	}

	go client.Write(r.Context())
	go client.Read(r.Context())
}
