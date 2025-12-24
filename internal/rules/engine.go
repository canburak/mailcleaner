package rules

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mailcleaner/mailcleaner/internal/config"
	"github.com/mailcleaner/mailcleaner/internal/imap"
)

// Engine handles rule execution
type Engine struct {
	config *config.Config
	dryRun bool
}

// NewEngine creates a new rule engine
func NewEngine(cfg *config.Config, dryRun bool) *Engine {
	return &Engine{
		config: cfg,
		dryRun: dryRun,
	}
}

// ExecuteAll executes all rules
func (e *Engine) ExecuteAll() error {
	log.Println("Starting rule execution...")

	for _, rule := range e.config.Rules {
		if err := e.ExecuteRule(&rule); err != nil {
			log.Printf("Error executing rule %s: %v", rule.Name, err)
			continue
		}
	}

	log.Println("Rule execution completed")
	return nil
}

// ExecuteRule executes a single rule
func (e *Engine) ExecuteRule(rule *config.Rule) error {
	log.Printf("Executing rule: %s", rule.Name)

	account := e.config.GetAccount(rule.Account)
	if account == nil {
		log.Printf("  Account not found: %s", rule.Account)
		return nil
	}

	client, err := imap.NewClient(account)
	if err != nil {
		return err
	}
	defer client.Close()

	mbox, err := client.SelectFolder(rule.Folder)
	if err != nil {
		return err
	}

	log.Printf("  Folder %s: %d messages", rule.Folder, mbox.Messages)

	if mbox.Messages == 0 {
		log.Println("  No messages to process")
		return nil
	}

	emails, err := client.FetchAllEmails(mbox)
	if err != nil {
		return err
	}

	matched := e.matchEmails(emails, &rule.Conditions)
	log.Printf("  Matched %d emails", len(matched))

	if len(matched) == 0 {
		return nil
	}

	return e.executeAction(client, matched, &rule.Action)
}

// matchEmails filters emails based on conditions
func (e *Engine) matchEmails(emails []*imap.Email, conditions *config.Conditions) []*imap.Email {
	var matched []*imap.Email

	for _, email := range emails {
		if e.matchConditions(email, conditions) {
			matched = append(matched, email)
		}
	}

	return matched
}

// matchConditions checks if an email matches all conditions
func (e *Engine) matchConditions(email *imap.Email, c *config.Conditions) bool {
	// From exact match
	if c.From != "" && !strings.EqualFold(email.From, c.From) {
		return false
	}

	// From contains
	if c.FromContains != "" && !strings.Contains(strings.ToLower(email.From), strings.ToLower(c.FromContains)) {
		return false
	}

	// To exact match
	if c.To != "" {
		found := false
		for _, to := range email.To {
			if strings.EqualFold(to, c.To) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// To contains
	if c.ToContains != "" {
		found := false
		for _, to := range email.To {
			if strings.Contains(strings.ToLower(to), strings.ToLower(c.ToContains)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Subject exact match
	if c.Subject != "" && !strings.EqualFold(email.Subject, c.Subject) {
		return false
	}

	// Subject contains
	if c.SubjectContains != "" && !strings.Contains(strings.ToLower(email.Subject), strings.ToLower(c.SubjectContains)) {
		return false
	}

	// Older than days
	if c.OlderThanDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -c.OlderThanDays)
		if email.Date.After(cutoff) {
			return false
		}
	}

	// Newer than days
	if c.NewerThanDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -c.NewerThanDays)
		if email.Date.Before(cutoff) {
			return false
		}
	}

	// Is read
	if c.IsRead != nil && email.IsRead() != *c.IsRead {
		return false
	}

	// Is unread (opposite of is_read)
	if c.IsUnread != nil && email.IsRead() == *c.IsUnread {
		return false
	}

	// Has attachment
	if c.HasAttachment != nil && email.HasAttachment != *c.HasAttachment {
		return false
	}

	// Size larger than
	if c.SizeLargerThan != "" {
		sizeBytes, err := parseSize(c.SizeLargerThan)
		if err == nil && email.Size <= sizeBytes {
			return false
		}
	}

	return true
}

// executeAction performs the action on matched emails
func (e *Engine) executeAction(client *imap.Client, emails []*imap.Email, action *config.Action) error {
	uids := make([]uint32, len(emails))
	for i, email := range emails {
		uids[i] = email.UID
		email.LogInfo()
	}

	if e.dryRun {
		log.Printf("  [DRY RUN] Would execute action: %s on %d emails", action.Type, len(emails))
		return nil
	}

	switch action.Type {
	case "delete":
		log.Printf("  Deleting %d emails", len(emails))
		if err := client.DeleteMessages(uids); err != nil {
			return err
		}
		return client.Expunge()

	case "move":
		if action.MoveTo == "" {
			log.Println("  Error: move action requires move_to destination")
			return nil
		}
		log.Printf("  Moving %d emails to %s", len(emails), action.MoveTo)
		return client.MoveMessages(uids, action.MoveTo)

	case "mark_read":
		log.Printf("  Marking %d emails as read", len(emails))
		return client.MarkAsRead(uids)

	case "mark_unread":
		log.Printf("  Marking %d emails as unread", len(emails))
		return client.MarkAsUnread(uids)

	case "flag":
		log.Printf("  Flagging %d emails", len(emails))
		return client.FlagMessages(uids)

	case "unflag":
		log.Printf("  Unflagging %d emails", len(emails))
		return client.UnflagMessages(uids)

	case "add_flag":
		if action.AddFlag == "" {
			log.Println("  Error: add_flag action requires add_flag value")
			return nil
		}
		log.Printf("  Adding flag %s to %d emails", action.AddFlag, len(emails))
		return client.SetFlags(uids, []string{action.AddFlag}, true)

	case "remove_flag":
		if action.RemoveFlag == "" {
			log.Println("  Error: remove_flag action requires remove_flag value")
			return nil
		}
		log.Printf("  Removing flag %s from %d emails", action.RemoveFlag, len(emails))
		return client.SetFlags(uids, []string{action.RemoveFlag}, false)

	default:
		log.Printf("  Unknown action type: %s", action.Type)
	}

	return nil
}

// parseSize parses a size string like "10MB" or "1GB" to bytes
func parseSize(s string) (uint32, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	multiplier := uint32(1)
	if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		s = strings.TrimSuffix(s, "B")
	}

	val, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(val) * multiplier, nil
}
