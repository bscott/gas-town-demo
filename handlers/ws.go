package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSMessage represents the WebSocket message format
type WSMessage struct {
	Type      string `json:"type"`
	ChannelID string `json:"channel_id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Client represents a WebSocket client connection
type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	channelID string
	hub       *Hub
}

// Hub maintains channel-specific client connections
type Hub struct {
	mu       sync.RWMutex
	channels map[string]map[*Client]bool
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]map[*Client]bool),
	}
}

// Register adds a client to a channel
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.channels[client.channelID] == nil {
		h.channels[client.channelID] = make(map[*Client]bool)
	}
	h.channels[client.channelID][client] = true
	log.Printf("Client connected to channel %s", client.channelID)
}

// Unregister removes a client from a channel
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.channels[client.channelID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
			log.Printf("Client disconnected from channel %s", client.channelID)
		}
		// Clean up empty channels
		if len(clients) == 0 {
			delete(h.channels, client.channelID)
		}
	}
}

// Broadcast sends a message to all clients in a channel
func (h *Hub) Broadcast(channelID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.channels[channelID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				// Client buffer full, skip
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the incoming message
		var msg WSMessage
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		// Ensure channel_id matches the client's channel
		msg.ChannelID = c.channelID
		msg.Type = "message"
		if msg.CreatedAt == "" {
			msg.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		}

		// Marshal and broadcast
		outMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		c.hub.Broadcast(c.channelID, outMsg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// WSHandler holds the WebSocket hub
type WSHandler struct {
	hub *Hub
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler() *WSHandler {
	return &WSHandler{
		hub: NewHub(),
	}
}

// HandleWebSocket handles WebSocket connections at /ws?channel=<id>
func (ws *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		http.Error(w, "channel parameter required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:      conn,
		send:      make(chan []byte, 256),
		channelID: channelID,
		hub:       ws.hub,
	}

	ws.hub.Register(client)

	go client.writePump()
	go client.readPump()
}

// RegisterRoutes registers the WebSocket route on the given mux
func (ws *WSHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws", ws.HandleWebSocket)
}
