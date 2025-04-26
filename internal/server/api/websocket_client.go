// Source: https://github.com/gorilla/websocket/blob/main/examples/chat/client.go

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marc921/talk/internal/types/openapi"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// TODO: ensure messages are sent by chunks and reconstructed on the client side or server database / file storage.
	maxMessageSize = 1 << 20 // 1 MB
)

// WebSocketClient is a middleman between the websocket connection and the hub.
type WebSocketClient struct {
	logger *zap.Logger
	hub    *WebSocketHub
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	out chan *openapi.Message
	// The username of the client.
	username openapi.Username
}

// newWebSocketClient creates a new WebSocketClient.
func NewWebSocketClient(
	logger *zap.Logger,
	hub *WebSocketHub,
	conn *websocket.Conn,
	username openapi.Username,
) *WebSocketClient {
	return &WebSocketClient{
		logger: logger.With(
			zap.String("component", "websocket_client"),
			zap.String("client_ip", conn.RemoteAddr().String()),
		),
		hub:      hub,
		conn:     conn,
		out:      make(chan *openapi.Message, 256),
		username: username,
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		// Extend the read deadline when a pong message is received.
		func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("conn.ReadMessage", zap.Error(err))
			}
			break
		}

		var msg *openapi.Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			c.logger.Error("json.Unmarshal", zap.Error(err))
			continue
		}

		if msg.Sender != c.username {
			c.logger.Warn("message sender does not match client username", zap.String("sender", string(msg.Sender)), zap.String("client_username", string(c.username)))
			continue
		}

		c.hub.in <- msg
	}
}

// WritePump pumps messages from the hub to the websocket connection.
//
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WebSocketClient) WritePump() error {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.out:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return nil
			}

			// Send the message to the client.
			msgBytes, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("json.Marshal: %w", err)
			}

			err = c.conn.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				return fmt.Errorf("conn.WriteMessage: %w", err)
			}
		case <-pingTicker.C:
			// Periodically send ping messages to the client to ensure they are still alive.
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return fmt.Errorf("conn.WriteMessage(Ping): %w", err)
			}
		}
	}
}
