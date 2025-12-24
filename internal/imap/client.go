// Package imap provides IMAP client functionality for mailcleaner
package imap

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"

	"github.com/mailcleaner/mailcleaner/internal/models"
)

// Client wraps the IMAP client with mailcleaner-specific functionality
type Client struct {
	conn     *client.Client
	account  *models.Account
	selected string
}

// Connect creates a new IMAP connection to the given account
func Connect(account *models.Account) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", account.Server, account.Port)

	var conn *client.Client
	var err error

	if account.TLS {
		conn, err = client.DialTLS(addr, nil)
	} else {
		conn, err = client.Dial(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", addr, err)
	}

	if err := conn.Login(account.Username, account.Password); err != nil {
		conn.Logout()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &Client{
		conn:    conn,
		account: account,
	}, nil
}

// Close logs out and closes the connection
func (c *Client) Close() error {
	return c.conn.Logout()
}

// TestConnection tests if the account credentials are valid
func (c *Client) TestConnection() (*models.ConnectionStatus, error) {
	status := &models.ConnectionStatus{Success: true, Message: "Connection successful"}

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 100)
	done := make(chan error, 1)

	go func() {
		done <- c.conn.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		status.Folders = append(status.Folders, models.Folder{
			Name:       m.Name,
			Delimiter:  m.Delimiter,
			Attributes: m.Attributes,
		})
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("listing mailboxes: %w", err)
	}

	// Get INBOX message count
	mbox, err := c.conn.Select("INBOX", true)
	if err != nil {
		return nil, fmt.Errorf("selecting INBOX: %w", err)
	}
	status.TotalEmails = int(mbox.Messages)

	return status, nil
}

// ListFolders returns all folders/mailboxes in the account
func (c *Client) ListFolders() ([]models.Folder, error) {
	mailboxes := make(chan *imap.MailboxInfo, 100)
	done := make(chan error, 1)

	go func() {
		done <- c.conn.List("", "*", mailboxes)
	}()

	var folders []models.Folder
	for m := range mailboxes {
		folders = append(folders, models.Folder{
			Name:       m.Name,
			Delimiter:  m.Delimiter,
			Attributes: m.Attributes,
		})
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("listing mailboxes: %w", err)
	}

	return folders, nil
}

// SelectFolder selects a mailbox/folder
func (c *Client) SelectFolder(name string) (int, error) {
	mbox, err := c.conn.Select(name, true)
	if err != nil {
		return 0, fmt.Errorf("selecting %s: %w", name, err)
	}
	c.selected = name
	return int(mbox.Messages), nil
}

// FetchMessages fetches messages from the currently selected folder
func (c *Client) FetchMessages(limit int) ([]models.Message, error) {
	if c.selected == "" {
		if _, err := c.SelectFolder("INBOX"); err != nil {
			return nil, err
		}
	}

	mbox, err := c.conn.Select(c.selected, true)
	if err != nil {
		return nil, fmt.Errorf("selecting %s: %w", c.selected, err)
	}

	if mbox.Messages == 0 {
		return []models.Message{}, nil
	}

	// Calculate range (fetch most recent messages first)
	from := uint32(1)
	to := mbox.Messages
	// Safe conversion: ensure limit is positive and within uint32 bounds
	if limit > 0 && limit <= int(^uint32(0)) {
		limitU32 := uint32(limit)
		if limitU32 < mbox.Messages {
			from = mbox.Messages - limitU32 + 1
		}
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	messages := make(chan *imap.Message, 100)
	done := make(chan error, 1)

	go func() {
		done <- c.conn.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchFlags}, messages)
	}()

	var result []models.Message
	for msg := range messages {
		if msg.Envelope == nil {
			continue
		}

		m := models.Message{
			UID:     msg.Uid,
			SeqNum:  msg.SeqNum,
			From:    formatAddresses(msg.Envelope.From),
			To:      formatAddresses(msg.Envelope.To),
			Subject: msg.Envelope.Subject,
			Date:    msg.Envelope.Date,
			Flags:   msg.Flags,
		}
		result = append(result, m)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("fetching messages: %w", err)
	}

	// Reverse to show most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// PreviewRules applies rules to messages and returns match results without moving
func (c *Client) PreviewRules(rules []models.Rule, folder string, limit int) (*models.PreviewResult, error) {
	if folder != "" {
		if _, err := c.SelectFolder(folder); err != nil {
			return nil, err
		}
	}

	messages, err := c.FetchMessages(limit)
	if err != nil {
		return nil, err
	}

	result := &models.PreviewResult{
		TotalMessages: len(messages),
		RuleMatches:   make(map[int64]int),
	}

	for i := range messages {
		msg := &messages[i]
		for j := range rules {
			rule := &rules[j]
			if !rule.Enabled {
				continue
			}

			if matchesRule(msg, rule) {
				msg.MatchedRule = rule
				result.MatchedMessages++
				result.RuleMatches[rule.ID]++
				break
			}
		}
	}

	result.Messages = messages
	return result, nil
}

// MoveMessage moves a message to a destination folder
func (c *Client) MoveMessage(uid uint32, destFolder string) error {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid)

	// Copy to destination folder
	if err := c.conn.UidCopy(seqSet, destFolder); err != nil {
		return fmt.Errorf("copying to %s: %w", destFolder, err)
	}

	// Mark original as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.conn.UidStore(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("marking as deleted: %w", err)
	}

	// Expunge deleted messages
	if err := c.conn.Expunge(nil); err != nil {
		return fmt.Errorf("expunging: %w", err)
	}

	return nil
}

// ApplyRules applies rules to messages and moves matching ones
func (c *Client) ApplyRules(rules []models.Rule, folder string, dryRun bool) (*models.PreviewResult, error) {
	preview, err := c.PreviewRules(rules, folder, 0)
	if err != nil {
		return nil, err
	}

	if dryRun {
		return preview, nil
	}

	for _, msg := range preview.Messages {
		if msg.MatchedRule != nil {
			if err := c.MoveMessage(msg.UID, msg.MatchedRule.MoveToFolder); err != nil {
				return nil, fmt.Errorf("moving message %d: %w", msg.UID, err)
			}
		}
	}

	return preview, nil
}

// CreateFolder creates a new folder/mailbox
func (c *Client) CreateFolder(name string) error {
	return c.conn.Create(name)
}

func matchesRule(msg *models.Message, rule *models.Rule) bool {
	pattern := strings.ToLower(rule.Pattern)

	switch rule.PatternType {
	case "sender", "":
		return strings.Contains(strings.ToLower(msg.From), pattern)
	case "subject":
		return strings.Contains(strings.ToLower(msg.Subject), pattern)
	case "from_domain":
		// Extract domain from From address
		from := strings.ToLower(msg.From)
		if idx := strings.LastIndex(from, "@"); idx != -1 {
			domain := from[idx+1:]
			// Remove trailing > if present
			domain = strings.TrimSuffix(domain, ">")
			return strings.Contains(domain, pattern)
		}
		return false
	default:
		return strings.Contains(strings.ToLower(msg.From), pattern)
	}
}

func formatAddresses(addresses []*imap.Address) string {
	var parts []string
	for _, addr := range addresses {
		if addr.PersonalName != "" {
			parts = append(parts, fmt.Sprintf("%s <%s@%s>", addr.PersonalName, addr.MailboxName, addr.HostName))
		} else {
			parts = append(parts, addr.MailboxName+"@"+addr.HostName)
		}
	}
	return strings.Join(parts, ", ")
}

// TestAccountConnection tests an account connection without keeping the client
func TestAccountConnection(account *models.Account) (*models.ConnectionStatus, error) {
	timeout := time.After(30 * time.Second)
	done := make(chan struct {
		status *models.ConnectionStatus
		err    error
	}, 1)

	go func() {
		client, err := Connect(account)
		if err != nil {
			done <- struct {
				status *models.ConnectionStatus
				err    error
			}{
				status: &models.ConnectionStatus{
					Success: false,
					Message: err.Error(),
				},
				err: nil,
			}
			return
		}
		defer client.Close()

		status, err := client.TestConnection()
		done <- struct {
			status *models.ConnectionStatus
			err    error
		}{status: status, err: err}
	}()

	select {
	case result := <-done:
		if result.err != nil {
			return &models.ConnectionStatus{
				Success: false,
				Message: result.err.Error(),
			}, nil
		}
		return result.status, nil
	case <-timeout:
		return &models.ConnectionStatus{
			Success: false,
			Message: "Connection timeout after 30 seconds",
		}, nil
	}
}
