package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WSMessage is a generic WebSocket event payload
type WSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// WSHub manages connected WebSocket clients
type WSHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ServeWS upgrades an HTTP connection to a WebSocket.
func (h *WSHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	log.Printf("📺 Dashboard client connected (%s)", r.RemoteAddr)

	// Keep-alive loop: discard incoming messages, remove on close
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
			log.Printf("📴 Dashboard client disconnected (%s)", r.RemoteAddr)
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

// Broadcast sends a message to all connected dashboard clients.
func (h *WSHub) Broadcast(msg WSMessage) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Printf("⚠️  WS write error: %v", err)
		}
	}
}

// ClientCount returns the number of active WebSocket connections.
func (h *WSHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
