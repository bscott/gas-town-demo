package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Channel represents a chat channel
type Channel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Message represents a message in a channel
type Message struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateChannelRequest is the request body for creating a channel
type CreateChannelRequest struct {
	Name string `json:"name"`
}

// CreateMessageRequest is the request body for sending a message
type CreateMessageRequest struct {
	Content string `json:"content"`
	Author  string `json:"author"`
}

// PaginatedMessages is the response for paginated message retrieval
type PaginatedMessages struct {
	Messages []Message `json:"messages"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
	Total    int       `json:"total"`
}

// API holds the state and handlers for the REST API
type API struct {
	mu          sync.RWMutex
	channels    map[string]*Channel
	messages    map[string][]Message
	channelSeq  int
	messageSeq  int
}

// NewAPI creates a new API instance
func NewAPI() *API {
	return &API{
		channels: make(map[string]*Channel),
		messages: make(map[string][]Message),
	}
}

// RegisterRoutes sets up the API routes on the given mux
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/channels", a.handleChannels)
	mux.HandleFunc("/api/channels/", a.handleChannelByID)
}

// handleChannels handles GET and POST /api/channels
func (a *API) handleChannels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listChannels(w, r)
	case http.MethodPost:
		a.createChannel(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleChannelByID routes requests for /api/channels/:id and /api/channels/:id/messages
func (a *API) handleChannelByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/channels/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Channel ID required", http.StatusBadRequest)
		return
	}

	channelID := parts[0]

	if len(parts) == 1 {
		// /api/channels/:id
		if r.Method == http.MethodGet {
			a.getChannel(w, r, channelID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) == 2 && parts[1] == "messages" {
		// /api/channels/:id/messages
		switch r.Method {
		case http.MethodGet:
			a.getMessages(w, r, channelID)
		case http.MethodPost:
			a.sendMessage(w, r, channelID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

// listChannels returns all channels
func (a *API) listChannels(w http.ResponseWriter, _ *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	channels := make([]Channel, 0, len(a.channels))
	for _, ch := range a.channels {
		channels = append(channels, *ch)
	}

	respondJSON(w, http.StatusOK, channels)
}

// createChannel creates a new channel
func (a *API) createChannel(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Channel name is required", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.channelSeq++
	channel := &Channel{
		ID:        strconv.Itoa(a.channelSeq),
		Name:      req.Name,
		CreatedAt: time.Now(),
	}
	a.channels[channel.ID] = channel
	a.messages[channel.ID] = []Message{}

	respondJSON(w, http.StatusCreated, channel)
}

// getChannel returns a single channel by ID
func (a *API) getChannel(w http.ResponseWriter, _ *http.Request, channelID string) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	channel, ok := a.channels[channelID]
	if !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, channel)
}

// getMessages returns messages for a channel with pagination
func (a *API) getMessages(w http.ResponseWriter, r *http.Request, channelID string) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if _, ok := a.channels[channelID]; !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	// Parse pagination params
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	messages := a.messages[channelID]
	total := len(messages)

	// Calculate pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		respondJSON(w, http.StatusOK, PaginatedMessages{
			Messages: []Message{},
			Page:     page,
			Limit:    limit,
			Total:    total,
		})
		return
	}

	if end > total {
		end = total
	}

	respondJSON(w, http.StatusOK, PaginatedMessages{
		Messages: messages[start:end],
		Page:     page,
		Limit:    limit,
		Total:    total,
	})
}

// sendMessage sends a message to a channel
func (a *API) sendMessage(w http.ResponseWriter, r *http.Request, channelID string) {
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Message content is required", http.StatusBadRequest)
		return
	}

	if req.Author == "" {
		http.Error(w, "Author is required", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.channels[channelID]; !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	a.messageSeq++
	message := Message{
		ID:        strconv.Itoa(a.messageSeq),
		ChannelID: channelID,
		Content:   req.Content,
		Author:    req.Author,
		CreatedAt: time.Now(),
	}
	a.messages[channelID] = append(a.messages[channelID], message)

	respondJSON(w, http.StatusCreated, message)
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
