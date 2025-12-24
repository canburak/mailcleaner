package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadValidConfig(t *testing.T) {
	content := `
accounts:
  - name: test
    host: imap.test.com
    port: 993
    username: user@test.com
    password: secret
    tls: true

rules:
  - name: test-rule
    account: test
    folder: INBOX
    conditions:
      from_contains: "@example.com"
    action:
      type: delete

schedule:
  interval_minutes: 30
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(cfg.Accounts))
	}
	if cfg.Accounts[0].Name != "test" {
		t.Errorf("expected account name 'test', got '%s'", cfg.Accounts[0].Name)
	}
	if cfg.Accounts[0].Port != 993 {
		t.Errorf("expected port 993, got %d", cfg.Accounts[0].Port)
	}

	if len(cfg.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Action.Type != "delete" {
		t.Errorf("expected action type 'delete', got '%s'", cfg.Rules[0].Action.Type)
	}

	if cfg.Schedule.IntervalMinutes != 30 {
		t.Errorf("expected interval 30, got %d", cfg.Schedule.IntervalMinutes)
	}
}

func TestLoadConfigWithDefaultPort(t *testing.T) {
	content := `
accounts:
  - name: test-tls
    host: imap.test.com
    username: user@test.com
    password: secret
    tls: true
  - name: test-plain
    host: imap.test.com
    username: user@test.com
    password: secret
    tls: false

rules:
  - name: test-rule
    account: test-tls
    folder: INBOX
    action:
      type: delete

schedule:
  interval_minutes: 60
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Accounts[0].Port != 993 {
		t.Errorf("expected TLS port 993, got %d", cfg.Accounts[0].Port)
	}
	if cfg.Accounts[1].Port != 143 {
		t.Errorf("expected plain port 143, got %d", cfg.Accounts[1].Port)
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "no accounts",
			content: "rules: []\nschedule:\n  interval_minutes: 60",
		},
		{
			name: "missing host",
			content: `
accounts:
  - name: test
    username: user@test.com
    password: secret
rules: []
`,
		},
		{
			name: "missing password",
			content: `
accounts:
  - name: test
    host: imap.test.com
    username: user@test.com
rules: []
`,
		},
		{
			name: "unknown account in rule",
			content: `
accounts:
  - name: test
    host: imap.test.com
    username: user@test.com
    password: secret
rules:
  - name: test-rule
    account: nonexistent
    folder: INBOX
    action:
      type: delete
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Error("Load() expected error, got nil")
			}
		})
	}
}

func TestAccountGetPassword(t *testing.T) {
	t.Run("direct password", func(t *testing.T) {
		acc := Account{Password: "secret123"}
		if acc.GetPassword() != "secret123" {
			t.Error("expected direct password")
		}
	})

	t.Run("env password", func(t *testing.T) {
		os.Setenv("TEST_PASSWORD", "env-secret")
		defer os.Unsetenv("TEST_PASSWORD")

		acc := Account{PasswordEnv: "TEST_PASSWORD"}
		if acc.GetPassword() != "env-secret" {
			t.Error("expected password from env")
		}
	})

	t.Run("env overrides direct", func(t *testing.T) {
		os.Setenv("TEST_PASSWORD", "env-secret")
		defer os.Unsetenv("TEST_PASSWORD")

		acc := Account{Password: "direct", PasswordEnv: "TEST_PASSWORD"}
		if acc.GetPassword() != "env-secret" {
			t.Error("expected env password to override direct")
		}
	})
}

func TestScheduleGetInterval(t *testing.T) {
	tests := []struct {
		minutes  int
		expected time.Duration
	}{
		{30, 30 * time.Minute},
		{60, 60 * time.Minute},
		{120, 120 * time.Minute},
		{0, 60 * time.Minute}, // default
	}

	for _, tt := range tests {
		s := Schedule{IntervalMinutes: tt.minutes}
		if s.GetInterval() != tt.expected {
			t.Errorf("GetInterval(%d) = %v, want %v", tt.minutes, s.GetInterval(), tt.expected)
		}
	}
}

func TestGetAccount(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Name: "personal"},
			{Name: "work"},
		},
	}

	if acc := cfg.GetAccount("personal"); acc == nil || acc.Name != "personal" {
		t.Error("expected to find 'personal' account")
	}
	if acc := cfg.GetAccount("work"); acc == nil || acc.Name != "work" {
		t.Error("expected to find 'work' account")
	}
	if acc := cfg.GetAccount("nonexistent"); acc != nil {
		t.Error("expected nil for nonexistent account")
	}
}
