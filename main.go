package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// Config holds the application configuration
type Config struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Rules    []Rule `json:"rules"`
}

// Rule defines a sender-matching rule
type Rule struct {
	Sender    string `json:"sender"`
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

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &config, nil
}

func run(config *Config, dryRun bool) error {
	// Connect to server
	addr := fmt.Sprintf("%s:%d", config.Server, config.Port)
	log.Printf("Connecting to %s...", addr)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("connecting to server: %w", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Username, config.Password); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	log.Println("Logged in successfully")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("selecting INBOX: %w", err)
	}
	log.Printf("INBOX has %d messages", mbox.Messages)

	if mbox.Messages == 0 {
		log.Println("No messages to process")
		return nil
	}

	// Fetch all messages
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 100)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid}, messages)
	}()

	// Process messages and apply rules
	var matches []matchedMessage
	for msg := range messages {
		if msg.Envelope == nil {
			continue
		}

		for _, rule := range config.Rules {
			if matchesSender(msg.Envelope.From, rule.Sender) {
				matches = append(matches, matchedMessage{
					UID:    msg.Uid,
					From:   formatAddresses(msg.Envelope.From),
					Subject: msg.Envelope.Subject,
					Rule:   rule,
				})
				break
			}
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("fetching messages: %w", err)
	}

	// Report and move matches
	log.Printf("Found %d matching messages", len(matches))

	for _, m := range matches {
		log.Printf("  %s -> %s (from: %s, subject: %s)",
			m.Rule.Sender, m.Rule.MoveToFolder, m.From, m.Subject)

		if !dryRun {
			if err := moveMessage(c, m.UID, m.Rule.MoveToFolder); err != nil {
				log.Printf("  ERROR moving message: %v", err)
			}
		}
	}

	if dryRun {
		log.Println("Dry run - no changes made")
	}

	return nil
}

type matchedMessage struct {
	UID     uint32
	From    string
	Subject string
	Rule    Rule
}

func matchesSender(addresses []*imap.Address, pattern string) bool {
	pattern = strings.ToLower(pattern)
	for _, addr := range addresses {
		email := strings.ToLower(addr.MailboxName + "@" + addr.HostName)
		if strings.Contains(email, pattern) {
			return true
		}
	}
	return false
}

func formatAddresses(addresses []*imap.Address) string {
	var parts []string
	for _, addr := range addresses {
		parts = append(parts, addr.MailboxName+"@"+addr.HostName)
	}
	return strings.Join(parts, ", ")
}

func moveMessage(c *client.Client, uid uint32, folder string) error {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid)

	// Copy to destination folder
	if err := c.UidCopy(seqSet, folder); err != nil {
		return fmt.Errorf("copying to %s: %w", folder, err)
	}

	// Mark original as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.UidStore(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("marking as deleted: %w", err)
	}

	// Expunge deleted messages
	if err := c.Expunge(nil); err != nil {
		return fmt.Errorf("expunging: %w", err)
	}

	return nil
}
