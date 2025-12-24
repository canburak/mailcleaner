package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `{
		"server": "imap.example.com",
		"port": 993,
		"username": "user@example.com",
		"password": "secret",
		"rules": [
			{"sender": "@newsletter.com", "move_to_folder": "Newsletters"},
			{"sender": "@github.com", "move_to_folder": "GitHub"}
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

	if config.Server != "imap.example.com" {
		t.Errorf("Server = %q, want %q", config.Server, "imap.example.com")
	}
	if config.Port != 993 {
		t.Errorf("Port = %d, want %d", config.Port, 993)
	}
	if config.Username != "user@example.com" {
		t.Errorf("Username = %q, want %q", config.Username, "user@example.com")
	}
	if config.Password != "secret" {
		t.Errorf("Password = %q, want %q", config.Password, "secret")
	}
	if len(config.Rules) != 2 {
		t.Errorf("Rules count = %d, want 2", len(config.Rules))
	}
	if config.Rules[0].Sender != "@newsletter.com" {
		t.Errorf("Rule[0].Sender = %q, want %q", config.Rules[0].Sender, "@newsletter.com")
	}
	if config.Rules[0].MoveToFolder != "Newsletters" {
		t.Errorf("Rule[0].MoveToFolder = %q, want %q", config.Rules[0].MoveToFolder, "Newsletters")
	}
}

func TestLoadConfigWithTLS(t *testing.T) {
	content := `{
		"server": "imap.example.com",
		"port": 143,
		"username": "user@example.com",
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

func TestLoadConfigTLSDefault(t *testing.T) {
	// When TLS is not specified, it should default to nil (interpreted as true)
	content := `{
		"server": "imap.example.com",
		"port": 993,
		"username": "user@example.com",
		"password": "secret",
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

	if config.TLS != nil {
		t.Errorf("TLS should be nil when not specified, got %v", *config.TLS)
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	_, err := loadConfig("/nonexistent/file.json")
	if err == nil {
		t.Error("loadConfig() should fail for nonexistent file")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	content := `{invalid json}`

	tmpfile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = loadConfig(tmpfile.Name())
	if err == nil {
		t.Error("loadConfig() should fail for invalid JSON")
	}
}
