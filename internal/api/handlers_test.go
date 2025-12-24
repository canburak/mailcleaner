package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/mailcleaner/mailcleaner/internal/models"
	"github.com/mailcleaner/mailcleaner/internal/storage"
)

func setupTestHandler(t *testing.T) (*Handler, *storage.Store, func()) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "mailcleaner-test-*.db")
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

	cleanup := func() {
		store.Close()
		os.Remove(tmpFile.Name())
	}

	return handler, store, cleanup
}

func TestListAccountsEmpty(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	w := httptest.NewRecorder()

	handler.ListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var accounts []models.Account
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if accounts != nil && len(accounts) != 0 {
		t.Errorf("Expected empty accounts, got %d", len(accounts))
	}
}

func TestCreateAccount(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	account := models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}

	body, _ := json.Marshal(account)
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateAccount(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.AccountWithoutPassword
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if created.Name != account.Name {
		t.Errorf("Expected name %s, got %s", account.Name, created.Name)
	}
	if created.ID == 0 {
		t.Error("Expected non-zero ID")
	}
}

func TestCreateAccountValidation(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Missing required fields
	account := models.Account{
		Name: "Test Account",
		// Missing server, username, password
	}

	body, _ := json.Marshal(account)
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetAccount(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create an account first
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create request with chi context
	req := httptest.NewRequest("GET", "/api/accounts/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var fetched models.AccountWithoutPassword
	if err := json.Unmarshal(w.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if fetched.Name != account.Name {
		t.Errorf("Expected name %s, got %s", account.Name, fetched.Name)
	}
}

func TestGetAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDeleteAccount(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create an account first
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/accounts/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.DeleteAccount(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify deletion
	accounts, _ := store.ListAccounts()
	if len(accounts) != 0 {
		t.Error("Account should have been deleted")
	}
}

func TestCreateRule(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account first
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	rule := models.Rule{
		Name:         "Newsletter Filter",
		Pattern:      "newsletter@",
		PatternType:  "sender",
		MoveToFolder: "Newsletters",
		Enabled:      true,
		Priority:     10,
	}

	body, _ := json.Marshal(rule)
	req := httptest.NewRequest("POST", "/api/accounts/1/rules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateRule(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.Rule
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if created.Name != rule.Name {
		t.Errorf("Expected name %s, got %s", rule.Name, created.Name)
	}
	if created.AccountID != 1 {
		t.Errorf("Expected account_id 1, got %d", created.AccountID)
	}
}

func TestListRules(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create rules
	for i := 0; i < 3; i++ {
		rule := &models.Rule{
			AccountID:    1,
			Name:         "Rule " + string(rune('A'+i)),
			Pattern:      "test",
			PatternType:  "sender",
			MoveToFolder: "Test",
			Enabled:      true,
			Priority:     i,
		}
		if err := store.CreateRule(rule); err != nil {
			t.Fatalf("Failed to create rule: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/api/accounts/1/rules", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ListRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var rules []models.Rule
	if err := json.Unmarshal(w.Body.Bytes(), &rules); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(rules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(rules))
	}
}

func TestUpdateRule(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account and rule
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	rule := &models.Rule{
		AccountID:    1,
		Name:         "Original Name",
		Pattern:      "test",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
		Priority:     0,
	}
	store.CreateRule(rule)

	// Update
	update := models.Rule{
		Name:         "Updated Name",
		Pattern:      "updated",
		PatternType:  "subject",
		MoveToFolder: "Updated",
		Enabled:      false,
		Priority:     5,
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/api/rules/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateRule(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Rule
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", updated.Name)
	}
	if updated.Pattern != "updated" {
		t.Errorf("Expected pattern 'updated', got %s", updated.Pattern)
	}
}

func TestDeleteRule(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account and rule
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	rule := &models.Rule{
		AccountID:    1,
		Name:         "To Delete",
		Pattern:      "test",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
		Priority:     0,
	}
	store.CreateRule(rule)

	req := httptest.NewRequest("DELETE", "/api/rules/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.DeleteRule(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify deletion
	rules, _ := store.ListRules(1)
	if len(rules) != 0 {
		t.Error("Rule should have been deleted")
	}
}
