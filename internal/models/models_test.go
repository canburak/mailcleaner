package models

import (
	"testing"
	"time"
)

func TestAccountToSafe(t *testing.T) {
	now := time.Now()
	account := &Account{
		ID:        1,
		Name:      "Test Account",
		Server:    "imap.example.com",
		Port:      993,
		Username:  "user@example.com",
		Password:  "secret123",
		TLS:       true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	safe := account.ToSafe()

	if safe.ID != account.ID {
		t.Errorf("ID mismatch: got %d, want %d", safe.ID, account.ID)
	}
	if safe.Name != account.Name {
		t.Errorf("Name mismatch: got %s, want %s", safe.Name, account.Name)
	}
	if safe.Server != account.Server {
		t.Errorf("Server mismatch: got %s, want %s", safe.Server, account.Server)
	}
	if safe.Port != account.Port {
		t.Errorf("Port mismatch: got %d, want %d", safe.Port, account.Port)
	}
	if safe.Username != account.Username {
		t.Errorf("Username mismatch: got %s, want %s", safe.Username, account.Username)
	}
	if safe.TLS != account.TLS {
		t.Errorf("TLS mismatch: got %v, want %v", safe.TLS, account.TLS)
	}
	if !safe.CreatedAt.Equal(account.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", safe.CreatedAt, account.CreatedAt)
	}
	if !safe.UpdatedAt.Equal(account.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", safe.UpdatedAt, account.UpdatedAt)
	}
}

func TestMessageMatchesRule(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		rule     Rule
		expected bool
	}{
		// Sender pattern type tests
		{
			name: "sender match exact email",
			message: Message{
				From: "newsletter@company.com",
			},
			rule: Rule{
				Pattern:     "newsletter@company.com",
				PatternType: "sender",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "sender match partial",
			message: Message{
				From: "John Doe <newsletter@company.com>",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "sender",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "sender match case insensitive",
			message: Message{
				From: "NEWSLETTER@COMPANY.COM",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "sender",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "sender no match",
			message: Message{
				From: "user@other.com",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "sender",
				Enabled:     true,
			},
			expected: false,
		},
		{
			name: "empty pattern type defaults to sender",
			message: Message{
				From: "newsletter@company.com",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "",
				Enabled:     true,
			},
			expected: true,
		},
		// Subject pattern type tests
		{
			name: "subject match",
			message: Message{
				From:    "user@example.com",
				Subject: "Weekly Newsletter - Issue 42",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "subject",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "subject match case insensitive",
			message: Message{
				From:    "user@example.com",
				Subject: "[URGENT] Action Required",
			},
			rule: Rule{
				Pattern:     "urgent",
				PatternType: "subject",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "subject no match",
			message: Message{
				From:    "user@example.com",
				Subject: "Hello World",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "subject",
				Enabled:     true,
			},
			expected: false,
		},
		// from_domain pattern type tests
		{
			name: "from_domain match",
			message: Message{
				From: "notifications@github.com",
			},
			rule: Rule{
				Pattern:     "github.com",
				PatternType: "from_domain",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "from_domain match with personal name",
			message: Message{
				From: "GitHub <notifications@github.com>",
			},
			rule: Rule{
				Pattern:     "github.com",
				PatternType: "from_domain",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "from_domain match partial",
			message: Message{
				From: "user@mail.company.com",
			},
			rule: Rule{
				Pattern:     "company.com",
				PatternType: "from_domain",
				Enabled:     true,
			},
			expected: true,
		},
		{
			name: "from_domain no match - domain in local part",
			message: Message{
				From: "github.com.user@other.com",
			},
			rule: Rule{
				Pattern:     "github.com",
				PatternType: "from_domain",
				Enabled:     true,
			},
			expected: false,
		},
		{
			name: "from_domain no match",
			message: Message{
				From: "user@example.com",
			},
			rule: Rule{
				Pattern:     "github.com",
				PatternType: "from_domain",
				Enabled:     true,
			},
			expected: false,
		},
		{
			name: "from_domain with no @ in address",
			message: Message{
				From: "invalid-email",
			},
			rule: Rule{
				Pattern:     "example.com",
				PatternType: "from_domain",
				Enabled:     true,
			},
			expected: false,
		},
		// Unknown pattern type defaults to sender
		{
			name: "unknown pattern type defaults to sender",
			message: Message{
				From: "newsletter@company.com",
			},
			rule: Rule{
				Pattern:     "newsletter",
				PatternType: "unknown_type",
				Enabled:     true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.MatchesRule(&tt.rule)
			if result != tt.expected {
				t.Errorf("MatchesRule() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesDomain(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		pattern  string
		expected bool
	}{
		{
			name:     "simple domain match",
			from:     "user@example.com",
			pattern:  "example.com",
			expected: true,
		},
		{
			name:     "subdomain match",
			from:     "user@mail.example.com",
			pattern:  "example.com",
			expected: true,
		},
		{
			name:     "with angle bracket",
			from:     "Name <user@example.com>",
			pattern:  "example.com",
			expected: true,
		},
		{
			name:     "case insensitive",
			from:     "user@EXAMPLE.COM",
			pattern:  "example.com",
			expected: true,
		},
		{
			name:     "no @ symbol",
			from:     "invalid",
			pattern:  "example.com",
			expected: false,
		},
		{
			name:     "domain not in host part",
			from:     "example.com@other.org",
			pattern:  "example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesDomain(tt.from, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesDomain(%q, %q) = %v, want %v", tt.from, tt.pattern, result, tt.expected)
			}
		})
	}
}
