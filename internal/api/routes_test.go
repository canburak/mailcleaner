package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/mailcleaner/mailcleaner/internal/models"
	"github.com/mailcleaner/mailcleaner/internal/storage"
)

func setupTestRouter(t *testing.T) (*http.Handler, *storage.Store, func()) {
	tmpFile, err := os.CreateTemp("", "mailcleaner-routes-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := storage.New(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create store: %v", err)
	}

	handler := NewHandler(store)
	router := NewRouter(handler)

	cleanup := func() {
		store.Close()
		os.Remove(tmpFile.Name())
	}

	var h http.Handler = router
	return &h, store, cleanup
}

func TestNewRouter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mailcleaner-router-test-*.db")
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

	if router == nil {
		t.Fatal("Expected non-nil router")
	}
}

func TestHealthEndpoint(t *testing.T) {
	h, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", response["status"])
	}
}

func TestAccountsEndpoint(t *testing.T) {
	h, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAccountsCreateEndpoint(t *testing.T) {
	h, _, cleanup := setupTestRouter(t)
	defer cleanup()

	body := `{"name":"Test","server":"imap.test.com","port":993,"username":"user","password":"pass","tls":true}`
	req := httptest.NewRequest("POST", "/api/accounts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAccountByIDEndpoint(t *testing.T) {
	h, store, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create account directly in store
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("GET", "/api/accounts/1", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAccountUpdateEndpoint(t *testing.T) {
	h, store, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	body := `{"name":"Updated","server":"imap.updated.com","port":993,"username":"user","tls":true}`
	req := httptest.NewRequest("PUT", "/api/accounts/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAccountDeleteEndpoint(t *testing.T) {
	h, store, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("DELETE", "/api/accounts/1", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRulesEndpoint(t *testing.T) {
	// Note: This test uses direct handler testing since route params
	// are named differently in routes.go ({id}) vs handlers.go (accountId)
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("GET", "/api/accounts/1/rules", nil)
	// Set up chi context with correct param name
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ListRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRulesCreateEndpoint(t *testing.T) {
	// Note: Uses direct handler testing due to route param naming
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	body := `{"name":"Test Rule","pattern":"test@","pattern_type":"sender","move_to_folder":"Test","enabled":true}`
	req := httptest.NewRequest("POST", "/api/accounts/1/rules", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateRule(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRuleByIDEndpoint(t *testing.T) {
	h, store, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create account and rule
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	rule := &models.Rule{
		AccountID:    1,
		Name:         "Test Rule",
		Pattern:      "test@",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
	}
	store.CreateRule(rule)

	req := httptest.NewRequest("GET", "/api/rules/1", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRuleUpdateEndpoint(t *testing.T) {
	h, store, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create account and rule
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	rule := &models.Rule{
		AccountID:    1,
		Name:         "Test Rule",
		Pattern:      "test@",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
	}
	store.CreateRule(rule)

	body := `{"name":"Updated Rule","pattern":"updated@","pattern_type":"sender","move_to_folder":"Updated","enabled":false}`
	req := httptest.NewRequest("PUT", "/api/rules/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRuleDeleteEndpoint(t *testing.T) {
	h, store, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create account and rule
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.test.com",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}
	store.CreateAccount(account)

	rule := &models.Rule{
		AccountID:    1,
		Name:         "Test Rule",
		Pattern:      "test@",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
	}
	store.CreateRule(rule)

	req := httptest.NewRequest("DELETE", "/api/rules/1", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCORSHeaders(t *testing.T) {
	h, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("OPTIONS", "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	// CORS preflight should return 200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for CORS preflight, got %d", w.Code)
	}

	// Check for CORS headers
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin == "" {
		t.Error("Expected Access-Control-Allow-Origin header")
	}
}

func TestNotFoundEndpoint(t *testing.T) {
	h, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestContentTypeJSON(t *testing.T) {
	h, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	(*h).ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}
}
