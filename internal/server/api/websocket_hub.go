// Source: https://github.com/gorilla/websocket/blob/main/examples/chat/hub.go

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/marc921/talk/internal/types/openapi"
	"go.uber.org/zap"
)

// WebSocketHub maintains the set of active clients and broadcasts messages to the clients.
type WebSocketHub struct {
	logger *zap.Logger
	// Registered clients.
	clients map[*WebSocketClient]bool
	// Inbound messages from the clients.
	in chan *openapi.Message
	// Register requests from the clients.
	register chan *WebSocketClient
	// Unregister requests from clients.
	unregister chan *WebSocketClient
	// Upgrader for the websocket connection.
	upgrader websocket.Upgrader
}

func NewWebSocketHub(logger *zap.Logger) *WebSocketHub {
	return &WebSocketHub{
		logger: logger.With(
			zap.String("component", "websocket_hub"),
		),
		in:         make(chan *openapi.Message),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		clients:    make(map[*WebSocketClient]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (h *WebSocketHub) close() {
	for client := range h.clients {
		h.unregisterClient(client)
	}
}

func (h *WebSocketHub) unregisterClient(client *WebSocketClient) {
	close(client.out)
	delete(h.clients, client)
}

func (h *WebSocketHub) Run(ctx context.Context) error {
	defer h.close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.unregisterClient(client)
			}
		case message := <-h.in:
			for client := range h.clients {
				if client.username != message.Recipient {
					continue
				}
				// TODO: persist messages in case user is not connected, deduplicate?
				select {
				case client.out <- message:
				default:
					h.unregisterClient(client)
				}
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func (h *WebSocketHub) RegisterClient(c echo.Context, username openapi.Username) error {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("upgrader.Upgrade: %w", err)
	}
	client := NewWebSocketClient(
		h.logger,
		h,
		conn,
		username,
	)
	h.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()

	return nil
}
