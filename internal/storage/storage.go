// Package storage provides SQLite-based persistence for accounts and rules
package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mailcleaner/mailcleaner/internal/models"
)

// Store handles all database operations
type Store struct {
	db *sql.DB
}

// New creates a new Store with the given database path
func New(dbPath string) (*Store, error) {
	// Enable foreign keys with connection string parameter
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			server TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 993,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			tls INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			pattern TEXT NOT NULL,
			pattern_type TEXT NOT NULL DEFAULT 'sender',
			move_to_folder TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			priority INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_account_id ON rules(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_priority ON rules(priority)`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("executing migration: %w", err)
		}
	}

	return nil
}

// Account Operations

// CreateAccount creates a new account
func (s *Store) CreateAccount(account *models.Account) error {
	now := time.Now()
	result, err := s.db.Exec(
		`INSERT INTO accounts (name, server, port, username, password, tls, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		account.Name, account.Server, account.Port, account.Username, account.Password,
		boolToInt(account.TLS), now, now,
	)
	if err != nil {
		return fmt.Errorf("inserting account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}

	account.ID = id
	account.CreatedAt = now
	account.UpdatedAt = now
	return nil
}

// GetAccount retrieves an account by ID
func (s *Store) GetAccount(id int64) (*models.Account, error) {
	account := &models.Account{}
	var tls int
	err := s.db.QueryRow(
		`SELECT id, name, server, port, username, password, tls, created_at, updated_at
		 FROM accounts WHERE id = ?`, id,
	).Scan(&account.ID, &account.Name, &account.Server, &account.Port,
		&account.Username, &account.Password, &tls,
		&account.CreatedAt, &account.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying account: %w", err)
	}
	account.TLS = intToBool(tls)
	return account, nil
}

// ListAccounts returns all accounts
func (s *Store) ListAccounts() ([]models.Account, error) {
	rows, err := s.db.Query(
		`SELECT id, name, server, port, username, password, tls, created_at, updated_at
		 FROM accounts ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		var tls int
		if err := rows.Scan(&account.ID, &account.Name, &account.Server, &account.Port,
			&account.Username, &account.Password, &tls,
			&account.CreatedAt, &account.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning account: %w", err)
		}
		account.TLS = intToBool(tls)
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

// UpdateAccount updates an existing account
func (s *Store) UpdateAccount(account *models.Account) error {
	account.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE accounts SET name = ?, server = ?, port = ?, username = ?, password = ?, tls = ?, updated_at = ?
		 WHERE id = ?`,
		account.Name, account.Server, account.Port, account.Username, account.Password,
		boolToInt(account.TLS), account.UpdatedAt, account.ID,
	)
	if err != nil {
		return fmt.Errorf("updating account: %w", err)
	}
	return nil
}

// DeleteAccount deletes an account and its associated rules
func (s *Store) DeleteAccount(id int64) error {
	_, err := s.db.Exec(`DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting account: %w", err)
	}
	return nil
}

// Rule Operations

// CreateRule creates a new rule
func (s *Store) CreateRule(rule *models.Rule) error {
	now := time.Now()
	result, err := s.db.Exec(
		`INSERT INTO rules (account_id, name, pattern, pattern_type, move_to_folder, enabled, priority, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.AccountID, rule.Name, rule.Pattern, rule.PatternType, rule.MoveToFolder,
		boolToInt(rule.Enabled), rule.Priority, now, now,
	)
	if err != nil {
		return fmt.Errorf("inserting rule: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}

	rule.ID = id
	rule.CreatedAt = now
	rule.UpdatedAt = now
	return nil
}

// GetRule retrieves a rule by ID
func (s *Store) GetRule(id int64) (*models.Rule, error) {
	rule := &models.Rule{}
	var enabled int
	err := s.db.QueryRow(
		`SELECT id, account_id, name, pattern, pattern_type, move_to_folder, enabled, priority, created_at, updated_at
		 FROM rules WHERE id = ?`, id,
	).Scan(&rule.ID, &rule.AccountID, &rule.Name, &rule.Pattern, &rule.PatternType,
		&rule.MoveToFolder, &enabled, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying rule: %w", err)
	}
	rule.Enabled = intToBool(enabled)
	return rule, nil
}

// ListRules returns all rules for an account
func (s *Store) ListRules(accountID int64) ([]models.Rule, error) {
	rows, err := s.db.Query(
		`SELECT id, account_id, name, pattern, pattern_type, move_to_folder, enabled, priority, created_at, updated_at
		 FROM rules WHERE account_id = ? ORDER BY priority DESC, name`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying rules: %w", err)
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var rule models.Rule
		var enabled int
		if err := rows.Scan(&rule.ID, &rule.AccountID, &rule.Name, &rule.Pattern, &rule.PatternType,
			&rule.MoveToFolder, &enabled, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning rule: %w", err)
		}
		rule.Enabled = intToBool(enabled)
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// ListAllRules returns all rules across all accounts
func (s *Store) ListAllRules() ([]models.Rule, error) {
	rows, err := s.db.Query(
		`SELECT id, account_id, name, pattern, pattern_type, move_to_folder, enabled, priority, created_at, updated_at
		 FROM rules ORDER BY account_id, priority DESC, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying rules: %w", err)
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var rule models.Rule
		var enabled int
		if err := rows.Scan(&rule.ID, &rule.AccountID, &rule.Name, &rule.Pattern, &rule.PatternType,
			&rule.MoveToFolder, &enabled, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning rule: %w", err)
		}
		rule.Enabled = intToBool(enabled)
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// UpdateRule updates an existing rule
func (s *Store) UpdateRule(rule *models.Rule) error {
	rule.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE rules SET account_id = ?, name = ?, pattern = ?, pattern_type = ?, move_to_folder = ?,
		 enabled = ?, priority = ?, updated_at = ? WHERE id = ?`,
		rule.AccountID, rule.Name, rule.Pattern, rule.PatternType, rule.MoveToFolder,
		boolToInt(rule.Enabled), rule.Priority, rule.UpdatedAt, rule.ID,
	)
	if err != nil {
		return fmt.Errorf("updating rule: %w", err)
	}
	return nil
}

// DeleteRule deletes a rule
func (s *Store) DeleteRule(id int64) error {
	_, err := s.db.Exec(`DELETE FROM rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting rule: %w", err)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
