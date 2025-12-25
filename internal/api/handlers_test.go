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

func TestUpdateAccount(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create an account first
	account := &models.Account{
		Name:     "Original Name",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Update account (without password - should keep existing)
	update := models.Account{
		Name:     "Updated Name",
		Server:   "imap.updated.com",
		Port:     143,
		Username: "new@example.com",
		TLS:      false,
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/api/accounts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.AccountWithoutPassword
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", updated.Name)
	}
	if updated.Server != "imap.updated.com" {
		t.Errorf("Expected server 'imap.updated.com', got %s", updated.Server)
	}

	// Verify password was kept
	fetched, _ := store.GetAccount(1)
	if fetched.Password != "password123" {
		t.Error("Password should have been preserved when not provided in update")
	}
}

func TestUpdateAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	update := models.Account{Name: "Test"}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest("PUT", "/api/accounts/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetRule(t *testing.T) {
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
		Name:         "Test Rule",
		Pattern:      "test@",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
		Priority:     5,
	}
	store.CreateRule(rule)

	req := httptest.NewRequest("GET", "/api/rules/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetRule(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var fetched models.Rule
	if err := json.Unmarshal(w.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if fetched.Name != rule.Name {
		t.Errorf("Expected name %s, got %s", rule.Name, fetched.Name)
	}
	if fetched.Pattern != rule.Pattern {
		t.Errorf("Expected pattern %s, got %s", rule.Pattern, fetched.Pattern)
	}
}

func TestGetRuleNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/rules/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetRule(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCreateAccountDefaultPort(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Account without port should default to 993
	account := models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Username: "test@example.com",
		Password: "password123",
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

	if created.Port != 993 {
		t.Errorf("Expected default port 993, got %d", created.Port)
	}
}

func TestCreateRuleValidation(t *testing.T) {
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
	store.CreateAccount(account)

	// Missing required fields
	rule := models.Rule{
		Name: "Test Rule",
		// Missing pattern and move_to_folder
	}

	body, _ := json.Marshal(rule)
	req := httptest.NewRequest("POST", "/api/accounts/1/rules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateRuleDefaultPatternType(t *testing.T) {
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
	store.CreateAccount(account)

	// Rule without pattern_type should default to "sender"
	rule := models.Rule{
		Name:         "Test Rule",
		Pattern:      "test@",
		MoveToFolder: "Test",
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

	if created.PatternType != "sender" {
		t.Errorf("Expected default pattern_type 'sender', got %s", created.PatternType)
	}
}

func TestGetAccountInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateRuleNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	update := models.Rule{Name: "Test"}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest("PUT", "/api/rules/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateRule(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestListAccountsWithData(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create some accounts
	for i := 0; i < 3; i++ {
		account := &models.Account{
			Name:     "Account " + string(rune('A'+i)),
			Server:   "imap.example.com",
			Port:     993,
			Username: "test@example.com",
			Password: "password123",
			TLS:      true,
		}
		store.CreateAccount(account)
	}

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	w := httptest.NewRecorder()

	handler.ListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var accounts []models.AccountWithoutPassword
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(accounts))
	}
}

// Tests for IMAP-dependent handlers

func TestTestAccountInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/accounts/invalid/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.TestAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTestAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/accounts/999/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.TestAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestTestAccountConnectionFailed(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account with invalid server
	account := &models.Account{
		Name:     "Test Account",
		Server:   "invalid.nonexistent.server",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("POST", "/api/accounts/1/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.TestAccount(w, req)

	// Should return 200 with failure status in body
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var status models.ConnectionStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if status.Success {
		t.Error("Expected connection to fail for invalid server")
	}
}

func TestTestAccountDirectInvalidBody(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/accounts/test", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.TestAccountDirect(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTestAccountDirectDefaultPort(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Account without port should default to 993
	account := models.Account{
		Name:     "Test",
		Server:   "invalid.nonexistent.server",
		Username: "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(account)
	req := httptest.NewRequest("POST", "/api/accounts/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.TestAccountDirect(w, req)

	// Should return 200 (connection test runs, returns result)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetAccountFoldersInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/invalid/folders", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetAccountFolders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetAccountFoldersNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/999/folders", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetAccountFolders(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetAccountFoldersConnectionFailed(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account with invalid server
	account := &models.Account{
		Name:     "Test Account",
		Server:   "invalid.nonexistent.server",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("GET", "/api/accounts/1/folders", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetAccountFolders(w, req)

	// Should return 502 Bad Gateway for connection failure
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPreviewRulesInvalidAccountID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/invalid/preview", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.PreviewRules(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPreviewRulesAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/999/preview", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.PreviewRules(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestPreviewRulesConnectionFailed(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account with invalid server
	account := &models.Account{
		Name:     "Test Account",
		Server:   "invalid.nonexistent.server",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("GET", "/api/accounts/1/preview?folder=INBOX&limit=10", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.PreviewRules(w, req)

	// Should return 502 Bad Gateway for connection failure
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPreviewRulesWithQueryParams(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test Account",
		Server:   "invalid.server",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	// Create a rule
	rule := &models.Rule{
		AccountID:    1,
		Name:         "Test Rule",
		Pattern:      "test@",
		PatternType:  "sender",
		MoveToFolder: "Test",
		Enabled:      true,
	}
	store.CreateRule(rule)

	req := httptest.NewRequest("GET", "/api/accounts/1/preview?folder=Spam&limit=50", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.PreviewRules(w, req)

	// Will fail due to connection, but tests the query param parsing
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502 (connection failed), got %d", w.Code)
	}
}

func TestApplyRulesInvalidAccountID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/accounts/invalid/apply", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ApplyRules(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestApplyRulesAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/accounts/999/apply", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ApplyRules(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestApplyRulesConnectionFailed(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account with invalid server
	account := &models.Account{
		Name:     "Test Account",
		Server:   "invalid.nonexistent.server",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("POST", "/api/accounts/1/apply?folder=INBOX&dry_run=true", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ApplyRules(w, req)

	// Should return 502 Bad Gateway for connection failure
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateFolderInvalidAccountID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	body := `{"name":"NewFolder"}`
	req := httptest.NewRequest("POST", "/api/accounts/invalid/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateFolder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateFolderAccountNotFound(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	body := `{"name":"NewFolder"}`
	req := httptest.NewRequest("POST", "/api/accounts/999/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateFolder(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCreateFolderInvalidBody(t *testing.T) {
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
	store.CreateAccount(account)

	req := httptest.NewRequest("POST", "/api/accounts/1/folders", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateFolder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateFolderEmptyName(t *testing.T) {
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
	store.CreateAccount(account)

	body := `{"name":""}`
	req := httptest.NewRequest("POST", "/api/accounts/1/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateFolder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateFolderConnectionFailed(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account with invalid server
	account := &models.Account{
		Name:     "Test Account",
		Server:   "invalid.nonexistent.server",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	body := `{"name":"NewFolder"}`
	req := httptest.NewRequest("POST", "/api/accounts/1/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateFolder(w, req)

	// Should return 502 Bad Gateway for connection failure
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteAccountInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/api/accounts/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.DeleteAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDeleteRuleInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/api/rules/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.DeleteRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestListRulesInvalidAccountID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/accounts/invalid/rules", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ListRules(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetRuleInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/rules/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateRuleInvalidAccountID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	body := `{"name":"Test","pattern":"test","move_to_folder":"Test"}`
	req := httptest.NewRequest("POST", "/api/accounts/invalid/rules", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateRuleInvalidBody(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("POST", "/api/accounts/1/rules", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.CreateRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateRuleInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	body := `{"name":"Test"}`
	req := httptest.NewRequest("PUT", "/api/rules/invalid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateRuleInvalidBody(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account and rule
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
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

	req := httptest.NewRequest("PUT", "/api/rules/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateRule(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateAccountInvalidID(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	body := `{"name":"Test"}`
	req := httptest.NewRequest("PUT", "/api/accounts/invalid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateAccountInvalidBody(t *testing.T) {
	handler, store, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create account
	account := &models.Account{
		Name:     "Test",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	req := httptest.NewRequest("PUT", "/api/accounts/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.UpdateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateAccountInvalidBody(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
