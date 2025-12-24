package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/emersion/go-imap"

	"github.com/mailcleaner/mailcleaner/testserver"
)

// parseServerAddr extracts host and port from a server address string
func parseServerAddr(addr string) (host string, port int) {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			host = addr[:i]
			fmt.Sscanf(addr[i+1:], "%d", &port)
			return
		}
	}
	return addr, 0
}

func TestMatchesSender(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []*imap.Address
		pattern  string
		expected bool
	}{
		{
			name: "exact domain match",
			addrs: []*imap.Address{
				{MailboxName: "news", HostName: "example.com"},
			},
			pattern:  "@example.com",
			expected: true,
		},
		{
			name: "partial match",
			addrs: []*imap.Address{
				{MailboxName: "newsletter", HostName: "company.com"},
			},
			pattern:  "newsletter",
			expected: true,
		},
		{
			name: "case insensitive",
			addrs: []*imap.Address{
				{MailboxName: "NOREPLY", HostName: "GitHub.com"},
			},
			pattern:  "github.com",
			expected: true,
		},
		{
			name: "no match",
			addrs: []*imap.Address{
				{MailboxName: "user", HostName: "other.com"},
			},
			pattern:  "example.com",
			expected: false,
		},
		{
			name:     "empty addresses",
			addrs:    []*imap.Address{},
			pattern:  "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSender(tt.addrs, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesSender(%v, %q) = %v, want %v",
					tt.addrs, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestFormatAddresses(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []*imap.Address
		expected string
	}{
		{
			name: "single address",
			addrs: []*imap.Address{
				{MailboxName: "test", HostName: "example.com"},
			},
			expected: "test@example.com",
		},
		{
			name: "multiple addresses",
			addrs: []*imap.Address{
				{MailboxName: "a", HostName: "x.com"},
				{MailboxName: "b", HostName: "y.com"},
			},
			expected: "a@x.com, b@y.com",
		},
		{
			name:     "empty",
			addrs:    []*imap.Address{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAddresses(tt.addrs)
			if result != tt.expected {
				t.Errorf("formatAddresses() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	content := `{
		"server": "test.example.com",
		"port": 993,
		"username": "user@test.com",
		"password": "secret",
		"rules": [
			{"sender": "@spam.com", "move_to_folder": "Spam"}
		]
	}`

	tmpfile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if config.Server != "test.example.com" {
		t.Errorf("Server = %q, want %q", config.Server, "test.example.com")
	}
	if config.Port != 993 {
		t.Errorf("Port = %d, want %d", config.Port, 993)
	}
	if len(config.Rules) != 1 {
		t.Errorf("Rules count = %d, want 1", len(config.Rules))
	}
	if config.Rules[0].Sender != "@spam.com" {
		t.Errorf("Rule sender = %q, want %q", config.Rules[0].Sender, "@spam.com")
	}
}

func TestLoadConfigWithTLS(t *testing.T) {
	content := `{
		"server": "test.example.com",
		"port": 143,
		"username": "user@test.com",
		"password": "secret",
		"tls": false,
		"rules": []
	}`

	tmpfile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if config.TLS == nil {
		t.Error("TLS should not be nil")
	} else if *config.TLS != false {
		t.Errorf("TLS = %v, want false", *config.TLS)
	}
}

// Integration tests using in-memory IMAP server

func TestIntegrationEmptyInbox(t *testing.T) {
	server, err := testserver.New("test@localhost", "password")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Close()

	host, port := parseServerAddr(server.Addr)

	tlsFalse := false
	config := &Config{
		Server:   host,
		Port:     port,
		Username: "test@localhost",
		Password: "password",
		TLS:      &tlsFalse,
		Rules: []Rule{
			{Sender: "@newsletter.com", MoveToFolder: "Newsletters"},
		},
	}

	// Should succeed with empty inbox
	err = run(config, true)
	if err != nil {
		t.Errorf("run() with empty inbox failed: %v", err)
	}
}

func TestIntegrationMatchAndDryRun(t *testing.T) {
	server, err := testserver.New("test@localhost", "password")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Close()

	// Add test messages
	server.AddMessage("newsletter@company.com", "Weekly News", "This is a newsletter")
	server.AddMessage("friend@gmail.com", "Hello", "Hi there!")
	server.AddMessage("sales@newsletter.com", "Special Offer", "Buy now!")

	host, port := parseServerAddr(server.Addr)

	tlsFalse := false
	config := &Config{
		Server:   host,
		Port:     port,
		Username: "test@localhost",
		Password: "password",
		TLS:      &tlsFalse,
		Rules: []Rule{
			{Sender: "newsletter", MoveToFolder: "Newsletters"},
		},
	}

	// Create destination folder
	server.CreateFolder("Newsletters")

	// Dry run should not move messages
	err = run(config, true)
	if err != nil {
		t.Errorf("run() dry run failed: %v", err)
	}

	// Messages should still be in INBOX
	inboxCount := server.GetMessageCount("INBOX")
	if inboxCount != 3 {
		t.Errorf("INBOX should have 3 messages after dry run, got %d", inboxCount)
	}
}

func TestIntegrationInvalidCredentials(t *testing.T) {
	server, err := testserver.New("test@localhost", "password")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Close()

	host, port := parseServerAddr(server.Addr)

	tlsFalse := false
	config := &Config{
		Server:   host,
		Port:     port,
		Username: "test@localhost",
		Password: "wrongpassword",
		TLS:      &tlsFalse,
		Rules:    []Rule{},
	}

	err = run(config, true)
	if err == nil {
		t.Error("run() with wrong password should fail")
	}
	if err.Error() != "login failed: invalid credentials" {
		t.Errorf("Expected login failure error, got: %v", err)
	}
}

func TestIntegrationNoMatchingRules(t *testing.T) {
	server, err := testserver.New("test@localhost", "password")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Close()

	server.AddMessage("friend@gmail.com", "Hello", "Hi there!")

	host, port := parseServerAddr(server.Addr)

	tlsFalse := false
	config := &Config{
		Server:   host,
		Port:     port,
		Username: "test@localhost",
		Password: "password",
		TLS:      &tlsFalse,
		Rules: []Rule{
			{Sender: "@spammer.com", MoveToFolder: "Spam"},
		},
	}

	err = run(config, true)
	if err != nil {
		t.Errorf("run() should succeed even with no matches: %v", err)
	}
}

func TestIntegrationMoveMessages(t *testing.T) {
	server, err := testserver.New("test@localhost", "password")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Close()

	// Add test messages
	server.AddMessage("newsletter@company.com", "Weekly News", "Newsletter content")
	server.AddMessage("friend@gmail.com", "Hello", "Hi there!")
	server.AddMessage("promo@newsletter.com", "Special Offer", "Buy now!")

	// Create destination folder
	server.CreateFolder("Newsletters")

	host, port := parseServerAddr(server.Addr)

	tlsFalse := false
	config := &Config{
		Server:   host,
		Port:     port,
		Username: "test@localhost",
		Password: "password",
		TLS:      &tlsFalse,
		Rules: []Rule{
			{Sender: "newsletter", MoveToFolder: "Newsletters"},
		},
	}

	// Run without dry-run - actually move messages
	err = run(config, false)
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	// Verify messages were moved
	inboxCount := server.GetMessageCount("INBOX")
	if inboxCount != 1 {
		t.Errorf("INBOX should have 1 message after move, got %d", inboxCount)
	}

	newsletterCount := server.GetMessageCount("Newsletters")
	if newsletterCount != 2 {
		t.Errorf("Newsletters should have 2 messages after move, got %d", newsletterCount)
	}
}
