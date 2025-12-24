// Package api provides the REST API handlers for mailcleaner
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	imapClient "github.com/mailcleaner/mailcleaner/internal/imap"
	"github.com/mailcleaner/mailcleaner/internal/models"
	"github.com/mailcleaner/mailcleaner/internal/storage"
)

// Handler holds dependencies for API handlers
type Handler struct {
	store *storage.Store
}

// NewHandler creates a new Handler
func NewHandler(store *storage.Store) *Handler {
	return &Handler{store: store}
}

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Account Handlers

// ListAccounts returns all accounts
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.store.ListAccounts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to safe format (no passwords)
	safeAccounts := make([]models.AccountWithoutPassword, len(accounts))
	for i, a := range accounts {
		safeAccounts[i] = a.ToSafe()
	}

	respondJSON(w, http.StatusOK, safeAccounts)
}

// GetAccount returns a single account
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	account, err := h.store.GetAccount(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if account == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	respondJSON(w, http.StatusOK, account.ToSafe())
}

// CreateAccount creates a new account
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var account models.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if account.Name == "" || account.Server == "" || account.Username == "" || account.Password == "" {
		respondError(w, http.StatusBadRequest, "name, server, username, and password are required")
		return
	}

	if account.Port == 0 {
		account.Port = 993
	}

	if err := h.store.CreateAccount(&account); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, account.ToSafe())
}

// UpdateAccount updates an existing account
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	existing, err := h.store.GetAccount(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	var account models.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account.ID = id
	// Keep existing password if not provided
	if account.Password == "" {
		account.Password = existing.Password
	}

	if err := h.store.UpdateAccount(&account); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, account.ToSafe())
}

// DeleteAccount deletes an account
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	if err := h.store.DeleteAccount(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// TestAccount tests the connection to an account
func (h *Handler) TestAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	account, err := h.store.GetAccount(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if account == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	status, err := imapClient.TestAccountConnection(account)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, status)
}

// TestAccountDirect tests a connection with provided credentials (no save)
func (h *Handler) TestAccountDirect(w http.ResponseWriter, r *http.Request) {
	var account models.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if account.Port == 0 {
		account.Port = 993
	}

	status, err := imapClient.TestAccountConnection(&account)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, status)
}

// GetAccountFolders returns all folders for an account
func (h *Handler) GetAccountFolders(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	account, err := h.store.GetAccount(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if account == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	client, err := imapClient.Connect(account)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer client.Close()

	folders, err := client.ListFolders()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, folders)
}

// Rule Handlers

// ListRules returns all rules for an account
func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(chi.URLParam(r, "accountId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	rules, err := h.store.ListRules(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rules)
}

// GetRule returns a single rule
func (h *Handler) GetRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid rule ID")
		return
	}

	rule, err := h.store.GetRule(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rule == nil {
		respondError(w, http.StatusNotFound, "rule not found")
		return
	}

	respondJSON(w, http.StatusOK, rule)
}

// CreateRule creates a new rule
func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(chi.URLParam(r, "accountId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule.AccountID = accountID

	if rule.Name == "" || rule.Pattern == "" || rule.MoveToFolder == "" {
		respondError(w, http.StatusBadRequest, "name, pattern, and move_to_folder are required")
		return
	}

	if rule.PatternType == "" {
		rule.PatternType = "sender"
	}

	if err := h.store.CreateRule(&rule); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, rule)
}

// UpdateRule updates an existing rule
func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid rule ID")
		return
	}

	existing, err := h.store.GetRule(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		respondError(w, http.StatusNotFound, "rule not found")
		return
	}

	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule.ID = id
	rule.AccountID = existing.AccountID

	if err := h.store.UpdateRule(&rule); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rule)
}

// DeleteRule deletes a rule
func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid rule ID")
		return
	}

	if err := h.store.DeleteRule(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// Preview Handler

// PreviewRules previews the effect of rules on an account's emails
func (h *Handler) PreviewRules(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(chi.URLParam(r, "accountId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	account, err := h.store.GetAccount(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if account == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	rules, err := h.store.ListRules(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get optional parameters
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "INBOX"
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	client, err := imapClient.Connect(account)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer client.Close()

	result, err := client.PreviewRules(rules, folder, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ApplyRules applies rules to move emails
func (h *Handler) ApplyRules(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(chi.URLParam(r, "accountId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	account, err := h.store.GetAccount(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if account == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	rules, err := h.store.ListRules(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "INBOX"
	}

	dryRun := r.URL.Query().Get("dry_run") == "true"

	client, err := imapClient.Connect(account)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer client.Close()

	result, err := client.ApplyRules(rules, folder, dryRun)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// CreateFolder creates a new folder in an account
func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(chi.URLParam(r, "accountId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	account, err := h.store.GetAccount(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if account == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "folder name is required")
		return
	}

	client, err := imapClient.Connect(account)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer client.Close()

	if err := client.CreateFolder(req.Name); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"name": req.Name})
}
