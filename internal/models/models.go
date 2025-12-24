// Package models defines the core data structures for the mailcleaner application
package models

import (
	"time"
)

// Account represents an IMAP email account
type Account struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Server    string    `json:"server"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	TLS       bool      `json:"tls"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountWithoutPassword is Account with password omitted for API responses
type AccountWithoutPassword struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Server    string    `json:"server"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	TLS       bool      `json:"tls"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToSafe converts an Account to AccountWithoutPassword
func (a *Account) ToSafe() AccountWithoutPassword {
	return AccountWithoutPassword{
		ID:        a.ID,
		Name:      a.Name,
		Server:    a.Server,
		Port:      a.Port,
		Username:  a.Username,
		TLS:       a.TLS,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// Rule defines a sender-matching rule for email organization
type Rule struct {
	ID           int64     `json:"id"`
	AccountID    int64     `json:"account_id"`
	Name         string    `json:"name"`
	Pattern      string    `json:"pattern"`
	PatternType  string    `json:"pattern_type"` // "sender", "subject", "from_domain"
	MoveToFolder string    `json:"move_to_folder"`
	Enabled      bool      `json:"enabled"`
	Priority     int       `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Message represents an email message for preview
type Message struct {
	UID         uint32    `json:"uid"`
	SeqNum      uint32    `json:"seq_num"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Subject     string    `json:"subject"`
	Date        time.Time `json:"date"`
	Flags       []string  `json:"flags"`
	MatchedRule *Rule     `json:"matched_rule,omitempty"`
}

// PreviewResult represents the result of applying rules to messages
type PreviewResult struct {
	TotalMessages   int              `json:"total_messages"`
	MatchedMessages int              `json:"matched_messages"`
	Messages        []Message        `json:"messages"`
	RuleMatches     map[int64]int    `json:"rule_matches"` // rule_id -> match count
}

// Folder represents an IMAP folder/mailbox
type Folder struct {
	Name       string   `json:"name"`
	Delimiter  string   `json:"delimiter"`
	Attributes []string `json:"attributes"`
}

// ConnectionStatus represents the status of an IMAP connection test
type ConnectionStatus struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	Folders     []Folder `json:"folders,omitempty"`
	TotalEmails int      `json:"total_emails,omitempty"`
}
