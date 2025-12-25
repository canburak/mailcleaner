package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mailcleaner/mailcleaner/internal/models"
	"github.com/mailcleaner/mailcleaner/internal/storage"
)

func setupTestWebSocket(t *testing.T) (*WebSocketHandler, *storage.Store, func()) {
	tmpFile, err := os.CreateTemp("", "mailcleaner-ws-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := storage.New(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create store: %v", err)
	}

	handler := NewWebSocketHandler(store)

	cleanup := func() {
		store.Close()
		os.Remove(tmpFile.Name())
	}

	return handler, store, cleanup
}

func TestNewWebSocketHandler(t *testing.T) {
	_, store, cleanup := setupTestWebSocket(t)
	defer cleanup()

	handler := NewWebSocketHandler(store)
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
	if handler.store != store {
		t.Error("Expected store to be set")
	}
}

func TestWSMessageSerialization(t *testing.T) {
	// Test WSMessage serialization
	msg := WSMessage{
		Type:  "test",
		Error: "test error",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Type != "test" {
		t.Errorf("Expected type 'test', got %s", decoded.Type)
	}
	if decoded.Error != "test error" {
		t.Errorf("Expected error 'test error', got %s", decoded.Error)
	}
}

func TestWSMessageWithPayload(t *testing.T) {
	payload := map[string]string{"key": "value"}
	payloadBytes, _ := json.Marshal(payload)

	msg := WSMessage{
		Type:    "data",
		Payload: payloadBytes,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Type != "data" {
		t.Errorf("Expected type 'data', got %s", decoded.Type)
	}

	var decodedPayload map[string]string
	if err := json.Unmarshal(decoded.Payload, &decodedPayload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if decodedPayload["key"] != "value" {
		t.Errorf("Expected payload key 'value', got %s", decodedPayload["key"])
	}
}

func TestPreviewRequestSerialization(t *testing.T) {
	req := PreviewRequest{
		AccountID: 123,
		Folder:    "INBOX",
		Limit:     50,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded PreviewRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.AccountID != 123 {
		t.Errorf("Expected AccountID 123, got %d", decoded.AccountID)
	}
	if decoded.Folder != "INBOX" {
		t.Errorf("Expected Folder 'INBOX', got %s", decoded.Folder)
	}
	if decoded.Limit != 50 {
		t.Errorf("Expected Limit 50, got %d", decoded.Limit)
	}
}

func TestPreviewProgressSerialization(t *testing.T) {
	progress := PreviewProgress{
		Stage:   "connecting",
		Current: 5,
		Total:   10,
		Message: "Processing...",
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded PreviewProgress
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Stage != "connecting" {
		t.Errorf("Expected Stage 'connecting', got %s", decoded.Stage)
	}
	if decoded.Current != 5 {
		t.Errorf("Expected Current 5, got %d", decoded.Current)
	}
	if decoded.Total != 10 {
		t.Errorf("Expected Total 10, got %d", decoded.Total)
	}
	if decoded.Message != "Processing..." {
		t.Errorf("Expected Message 'Processing...', got %s", decoded.Message)
	}
}

func TestPreviewProgressWithMessageData(t *testing.T) {
	msg := &models.Message{
		UID:     1,
		From:    "test@example.com",
		Subject: "Test Subject",
	}

	progress := PreviewProgress{
		Stage:       "processing",
		Current:     1,
		Total:       1,
		Message:     "Done",
		MessageData: msg,
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded PreviewProgress
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.MessageData == nil {
		t.Fatal("Expected MessageData to be set")
	}
	if decoded.MessageData.UID != 1 {
		t.Errorf("Expected UID 1, got %d", decoded.MessageData.UID)
	}
	if decoded.MessageData.From != "test@example.com" {
		t.Errorf("Expected From 'test@example.com', got %s", decoded.MessageData.From)
	}
}

func TestHandleLivePreviewUpgrade(t *testing.T) {
	handler, _, cleanup := setupTestWebSocket(t)
	defer cleanup()

	// Create test server with WebSocket handler
	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status 101, got %d", resp.StatusCode)
	}
}

func TestHandleLivePreviewPing(t *testing.T) {
	handler, _, cleanup := setupTestWebSocket(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	// Send ping message
	pingMsg := WSMessage{Type: "ping"}
	if err := conn.WriteJSON(pingMsg); err != nil {
		t.Fatalf("Failed to write ping: %v", err)
	}

	// Expect pong response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response WSMessage
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if response.Type != "pong" {
		t.Errorf("Expected pong response, got %s", response.Type)
	}
}

func TestHandleLivePreviewUnknownType(t *testing.T) {
	handler, _, cleanup := setupTestWebSocket(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	// Send unknown message type
	unknownMsg := WSMessage{Type: "unknown_type"}
	if err := conn.WriteJSON(unknownMsg); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// Expect error response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response WSMessage
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if response.Type != "error" {
		t.Errorf("Expected error response, got %s", response.Type)
	}
	if response.Error != "unknown message type" {
		t.Errorf("Expected 'unknown message type' error, got %s", response.Error)
	}
}

func TestHandleLivePreviewInvalidRequest(t *testing.T) {
	handler, _, cleanup := setupTestWebSocket(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	// Send preview request with invalid payload (missing required fields in JSON)
	// Using a JSON object that won't parse to PreviewRequest properly
	msg := WSMessage{
		Type:    "preview",
		Payload: json.RawMessage(`{"invalid": true, "bad_field": "not_a_number"}`),
	}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// Read responses - we should get progress then error for account not found
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response WSMessage
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Should either be an error or progress followed by error
	if response.Type != "progress" && response.Type != "error" {
		t.Errorf("Expected progress or error response, got %s", response.Type)
	}
}

func TestHandleLivePreviewAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestWebSocket(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	// Send preview request for non-existent account
	payload, _ := json.Marshal(PreviewRequest{
		AccountID: 999,
		Folder:    "INBOX",
		Limit:     10,
	})
	msg := WSMessage{
		Type:    "preview",
		Payload: payload,
	}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// First message should be progress (connecting)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response WSMessage
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Check if progress or error
	if response.Type == "progress" {
		// Next should be error for account not found
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatalf("Failed to read second response: %v", err)
		}
	}

	if response.Type != "error" {
		t.Errorf("Expected error for non-existent account, got %s", response.Type)
	}
	if response.Error != "account not found" {
		t.Errorf("Expected 'account not found' error, got %s", response.Error)
	}
}

func TestHandleLivePreviewDefaultValues(t *testing.T) {
	handler, store, cleanup := setupTestWebSocket(t)
	defer cleanup()

	// Create an account
	account := &models.Account{
		Name:     "Test",
		Server:   "invalid.server",
		Port:     993,
		Username: "test",
		Password: "test",
		TLS:      true,
	}
	store.CreateAccount(account)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	// Send preview request with no folder/limit (should use defaults)
	payload, _ := json.Marshal(PreviewRequest{
		AccountID: 1,
		// No Folder or Limit specified
	})
	msg := WSMessage{
		Type:    "preview",
		Payload: payload,
	}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// Read response (will be progress then error due to invalid server)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response WSMessage
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Should get a progress message first
	if response.Type != "progress" {
		t.Errorf("Expected progress message first, got %s", response.Type)
	}
}

func TestAddWebSocketRoutes(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mailcleaner-ws-routes-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := storage.New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	handler := NewHandler(store)
	router := NewRouter(handler)

	// Add WebSocket routes
	AddWebSocketRoutes(router, store)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test that WebSocket endpoint exists
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/preview"

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status 101, got %d", resp.StatusCode)
	}
}

func TestUpgraderCheckOrigin(t *testing.T) {
	// Test that the upgrader allows all origins (for development)
	req := httptest.NewRequest("GET", "/ws/preview", nil)
	req.Header.Set("Origin", "http://someorigin.com")

	// The upgrader.CheckOrigin should return true for all origins
	result := upgrader.CheckOrigin(req)
	if !result {
		t.Error("Expected CheckOrigin to return true for any origin")
	}
}

func TestConnectionClose(t *testing.T) {
	handler, _, cleanup := setupTestWebSocket(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleLivePreview))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}

	// Close connection normally
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		t.Logf("Write close message: %v", err)
	}

	conn.Close()
}
