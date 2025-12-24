// Package main provides the CLI for mailcleaner
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	imapClient "github.com/mailcleaner/mailcleaner/internal/imap"
	"github.com/mailcleaner/mailcleaner/internal/models"
)

// LegacyConfig holds the legacy configuration format for backwards compatibility
type LegacyConfig struct {
	Server   string       `json:"server"`
	Port     int          `json:"port"`
	Username string       `json:"username"`
	Password string       `json:"password"`
	TLS      *bool        `json:"tls,omitempty"`
	Rules    []LegacyRule `json:"rules"`
}

// LegacyRule defines the legacy rule format
type LegacyRule struct {
	Sender       string `json:"sender"`
	MoveToFolder string `json:"move_to_folder"`
}

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	dryRun := flag.Bool("dry-run", false, "show what would be done without making changes")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := run(config, *dryRun); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func loadConfig(path string) (*LegacyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config LegacyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &config, nil
}

func run(config *LegacyConfig, dryRun bool) error {
	// Convert legacy config to new models
	useTLS := config.TLS == nil || *config.TLS
	account := &models.Account{
		Server:   config.Server,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		TLS:      useTLS,
	}

	var rules []models.Rule
	for i, r := range config.Rules {
		rules = append(rules, models.Rule{
			ID:           int64(i + 1),
			Name:         fmt.Sprintf("Rule %d: %s", i+1, r.Sender),
			Pattern:      r.Sender,
			PatternType:  "sender",
			MoveToFolder: r.MoveToFolder,
			Enabled:      true,
			Priority:     len(config.Rules) - i, // Higher priority for earlier rules
		})
	}

	// Connect to IMAP server
	addr := fmt.Sprintf("%s:%d", account.Server, account.Port)
	log.Printf("Connecting to %s...", addr)

	client, err := imapClient.Connect(account)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer client.Close()

	log.Println("Logged in successfully")

	// Apply rules
	result, err := client.ApplyRules(rules, "INBOX", dryRun)
	if err != nil {
		return fmt.Errorf("applying rules: %w", err)
	}

	log.Printf("Processed %d messages, %d matched rules", result.TotalMessages, result.MatchedMessages)

	for _, msg := range result.Messages {
		if msg.MatchedRule != nil {
			log.Printf("  %s -> %s (from: %s, subject: %s)",
				msg.MatchedRule.Pattern, msg.MatchedRule.MoveToFolder, msg.From, msg.Subject)
		}
	}

	if dryRun {
		log.Println("Dry run - no changes made")
	}

	return nil
}
