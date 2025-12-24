package storage

import (
	"os"
	"testing"

	"github.com/mailcleaner/mailcleaner/internal/models"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	tmpFile, err := os.CreateTemp("", "mailcleaner-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := New(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(tmpFile.Name())
	}

	return store, cleanup
}

func TestAccountCRUD(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}

	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	if account.ID == 0 {
		t.Error("Expected non-zero ID after create")
	}

	// Read
	fetched, err := store.GetAccount(account.ID)
	if err != nil {
		t.Fatalf("GetAccount failed: %v", err)
	}

	if fetched.Name != account.Name {
		t.Errorf("Expected name %s, got %s", account.Name, fetched.Name)
	}

	// Update
	account.Name = "Updated Account"
	if err := store.UpdateAccount(account); err != nil {
		t.Fatalf("UpdateAccount failed: %v", err)
	}

	fetched, _ = store.GetAccount(account.ID)
	if fetched.Name != "Updated Account" {
		t.Errorf("Expected name 'Updated Account', got %s", fetched.Name)
	}

	// List
	accounts, err := store.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(accounts))
	}

	// Delete
	if err := store.DeleteAccount(account.ID); err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}

	fetched, _ = store.GetAccount(account.ID)
	if fetched != nil {
		t.Error("Account should have been deleted")
	}
}

func TestRuleCRUD(t *testing.T) {
	store, cleanup := setupTestStore(t)
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

	// Create rule
	rule := &models.Rule{
		AccountID:    account.ID,
		Name:         "Test Rule",
		Pattern:      "newsletter@",
		PatternType:  "sender",
		MoveToFolder: "Newsletters",
		Enabled:      true,
		Priority:     10,
	}

	if err := store.CreateRule(rule); err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	if rule.ID == 0 {
		t.Error("Expected non-zero ID after create")
	}

	// Read
	fetched, err := store.GetRule(rule.ID)
	if err != nil {
		t.Fatalf("GetRule failed: %v", err)
	}

	if fetched.Name != rule.Name {
		t.Errorf("Expected name %s, got %s", rule.Name, fetched.Name)
	}

	// Update
	rule.Name = "Updated Rule"
	rule.Priority = 20
	if err := store.UpdateRule(rule); err != nil {
		t.Fatalf("UpdateRule failed: %v", err)
	}

	fetched, _ = store.GetRule(rule.ID)
	if fetched.Name != "Updated Rule" {
		t.Errorf("Expected name 'Updated Rule', got %s", fetched.Name)
	}
	if fetched.Priority != 20 {
		t.Errorf("Expected priority 20, got %d", fetched.Priority)
	}

	// List
	rules, err := store.ListRules(account.ID)
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	// Delete
	if err := store.DeleteRule(rule.ID); err != nil {
		t.Fatalf("DeleteRule failed: %v", err)
	}

	fetched, _ = store.GetRule(rule.ID)
	if fetched != nil {
		t.Error("Rule should have been deleted")
	}
}

func TestRulePrioritySorting(t *testing.T) {
	store, cleanup := setupTestStore(t)
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

	// Create rules with different priorities
	priorities := []int{5, 10, 1, 8}
	for i, p := range priorities {
		rule := &models.Rule{
			AccountID:    account.ID,
			Name:         "Rule " + string(rune('A'+i)),
			Pattern:      "test",
			PatternType:  "sender",
			MoveToFolder: "Test",
			Enabled:      true,
			Priority:     p,
		}
		store.CreateRule(rule)
	}

	rules, _ := store.ListRules(account.ID)

	// Should be sorted by priority descending
	for i := 0; i < len(rules)-1; i++ {
		if rules[i].Priority < rules[i+1].Priority {
			t.Errorf("Rules not sorted by priority: %d < %d", rules[i].Priority, rules[i+1].Priority)
		}
	}
}

func TestCascadeDeleteRules(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create account with rules
	account := &models.Account{
		Name:     "Test Account",
		Server:   "imap.example.com",
		Port:     993,
		Username: "test@example.com",
		Password: "password123",
		TLS:      true,
	}
	store.CreateAccount(account)

	for i := 0; i < 3; i++ {
		rule := &models.Rule{
			AccountID:    account.ID,
			Name:         "Rule " + string(rune('A'+i)),
			Pattern:      "test",
			PatternType:  "sender",
			MoveToFolder: "Test",
			Enabled:      true,
			Priority:     i,
		}
		store.CreateRule(rule)
	}

	// Verify rules exist
	rules, _ := store.ListRules(account.ID)
	if len(rules) != 3 {
		t.Fatalf("Expected 3 rules, got %d", len(rules))
	}

	// Delete account - should cascade delete rules
	store.DeleteAccount(account.ID)

	// Verify rules are deleted
	rules, _ = store.ListRules(account.ID)
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules after account deletion, got %d", len(rules))
	}
}
