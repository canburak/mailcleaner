package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	imapClient "github.com/mailcleaner/mailcleaner/internal/imap"
	"github.com/mailcleaner/mailcleaner/internal/models"
	"github.com/mailcleaner/mailcleaner/internal/storage"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketHandler handles WebSocket connections for live preview
type WebSocketHandler struct {
	store *storage.Store
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(store *storage.Store) *WebSocketHandler {
	return &WebSocketHandler{store: store}
}

// Message types for WebSocket communication
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type PreviewRequest struct {
	AccountID int64  `json:"account_id"`
	Folder    string `json:"folder"`
	Limit     int    `json:"limit"`
}

type PreviewProgress struct {
	Stage       string          `json:"stage"`
	Current     int             `json:"current"`
	Total       int             `json:"total"`
	Message     string          `json:"message"`
	MessageData *models.Message `json:"message_data,omitempty"`
}

// HandleLivePreview handles WebSocket connections for live email preview
func (h *WebSocketHandler) HandleLivePreview(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Set up ping/pong for connection health
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Handle messages
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "preview":
			h.handlePreviewRequest(conn, msg.Payload)
		case "ping":
			conn.WriteJSON(WSMessage{Type: "pong"})
		default:
			conn.WriteJSON(WSMessage{Type: "error", Error: "unknown message type"})
		}
	}
}

func (h *WebSocketHandler) handlePreviewRequest(conn *websocket.Conn, payload json.RawMessage) {
	var req PreviewRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		conn.WriteJSON(WSMessage{Type: "error", Error: "invalid preview request"})
		return
	}

	if req.Folder == "" {
		req.Folder = "INBOX"
	}
	if req.Limit == 0 {
		req.Limit = 100
	}

	// Send connecting status
	h.sendProgress(conn, "connecting", 0, 0, "Connecting to IMAP server...")

	account, err := h.store.GetAccount(req.AccountID)
	if err != nil || account == nil {
		conn.WriteJSON(WSMessage{Type: "error", Error: "account not found"})
		return
	}

	rules, err := h.store.ListRules(req.AccountID)
	if err != nil {
		conn.WriteJSON(WSMessage{Type: "error", Error: "failed to load rules"})
		return
	}

	client, err := imapClient.Connect(account)
	if err != nil {
		conn.WriteJSON(WSMessage{Type: "error", Error: err.Error()})
		return
	}
	defer client.Close()

	h.sendProgress(conn, "connected", 0, 0, "Connected successfully")

	// Select folder
	h.sendProgress(conn, "selecting", 0, 0, "Selecting folder: "+req.Folder)
	totalMessages, err := client.SelectFolder(req.Folder)
	if err != nil {
		conn.WriteJSON(WSMessage{Type: "error", Error: err.Error()})
		return
	}

	h.sendProgress(conn, "fetching", 0, totalMessages, "Fetching messages...")

	// Fetch messages
	messages, err := client.FetchMessages(req.Limit)
	if err != nil {
		conn.WriteJSON(WSMessage{Type: "error", Error: err.Error()})
		return
	}

	h.sendProgress(conn, "processing", 0, len(messages), "Processing rules...")

	// Apply rules and send progress for each message
	result := &models.PreviewResult{
		TotalMessages: len(messages),
		RuleMatches:   make(map[int64]int),
	}

	for i := range messages {
		msg := &messages[i]

		for j := range rules {
			rule := &rules[j]
			if !rule.Enabled {
				continue
			}

			if msg.MatchesRule(rule) {
				msg.MatchedRule = rule
				result.MatchedMessages++
				result.RuleMatches[rule.ID]++
				break
			}
		}

		// Send progress update with message data
		h.sendProgressWithMessage(conn, "processing", i+1, len(messages),
			"Processing message "+strconv.Itoa(i+1)+" of "+strconv.Itoa(len(messages)), msg)
	}

	result.Messages = messages

	// Send final result
	resultData, _ := json.Marshal(result)
	conn.WriteJSON(WSMessage{Type: "result", Payload: resultData})
}

func (h *WebSocketHandler) sendProgress(conn *websocket.Conn, stage string, current, total int, message string) {
	progress := PreviewProgress{
		Stage:   stage,
		Current: current,
		Total:   total,
		Message: message,
	}
	data, _ := json.Marshal(progress)
	conn.WriteJSON(WSMessage{Type: "progress", Payload: data})
}

func (h *WebSocketHandler) sendProgressWithMessage(conn *websocket.Conn, stage string, current, total int, message string, msgData *models.Message) {
	progress := PreviewProgress{
		Stage:       stage,
		Current:     current,
		Total:       total,
		Message:     message,
		MessageData: msgData,
	}
	data, _ := json.Marshal(progress)
	conn.WriteJSON(WSMessage{Type: "progress", Payload: data})
}

// AddWebSocketRoutes adds WebSocket routes to the router
func AddWebSocketRoutes(r *chi.Mux, store *storage.Store) {
	wsHandler := NewWebSocketHandler(store)
	r.Get("/ws/preview", wsHandler.HandleLivePreview)
}
