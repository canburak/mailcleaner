package imap

import (
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/mailcleaner/mailcleaner/internal/config"
)

// Client wraps the IMAP client with convenience methods
type Client struct {
	client  *client.Client
	account *config.Account
}

// Email represents a fetched email message
type Email struct {
	UID         uint32
	SeqNum      uint32
	From        string
	To          []string
	Subject     string
	Date        time.Time
	Flags       []string
	Size        uint32
	HasAttachment bool
}

// NewClient creates a new IMAP client connection
func NewClient(account *config.Account) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", account.Host, account.Port)

	var c *client.Client
	var err error

	if account.TLS {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	if err := c.Login(account.Username, account.GetPassword()); err != nil {
		c.Logout()
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &Client{
		client:  c,
		account: account,
	}, nil
}

// Close logs out and closes the connection
func (c *Client) Close() error {
	return c.client.Logout()
}

// ListFolders returns all available mailbox folders
func (c *Client) ListFolders() ([]string, error) {
	mailboxes := make(chan *imap.MailboxInfo, 100)
	done := make(chan error, 1)

	go func() {
		done <- c.client.List("", "*", mailboxes)
	}()

	var folders []string
	for m := range mailboxes {
		folders = append(folders, m.Name)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	return folders, nil
}

// SelectFolder selects a mailbox folder
func (c *Client) SelectFolder(folder string) (*imap.MailboxStatus, error) {
	mbox, err := c.client.Select(folder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select folder %s: %w", folder, err)
	}
	return mbox, nil
}

// FetchEmails fetches emails from the currently selected folder
func (c *Client) FetchEmails(seqSet *imap.SeqSet) ([]*Email, error) {
	if seqSet.Empty() {
		return nil, nil
	}

	messages := make(chan *imap.Message, 100)
	done := make(chan error, 1)

	section := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchFlags,
		imap.FetchUid,
		imap.FetchRFC822Size,
		imap.FetchBodyStructure,
		section.FetchItem(),
	}

	go func() {
		done <- c.client.Fetch(seqSet, items, messages)
	}()

	var emails []*Email
	for msg := range messages {
		email := &Email{
			UID:    msg.Uid,
			SeqNum: msg.SeqNum,
			Flags:  msg.Flags,
			Size:   msg.Size,
		}

		if msg.Envelope != nil {
			email.Subject = msg.Envelope.Subject
			email.Date = msg.Envelope.Date

			if len(msg.Envelope.From) > 0 {
				from := msg.Envelope.From[0]
				email.From = fmt.Sprintf("%s@%s", from.MailboxName, from.HostName)
			}

			for _, addr := range msg.Envelope.To {
				email.To = append(email.To, fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName))
			}
		}

		if msg.BodyStructure != nil {
			email.HasAttachment = hasAttachment(msg.BodyStructure)
		}

		emails = append(emails, email)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch emails: %w", err)
	}

	return emails, nil
}

// FetchAllEmails fetches all emails from the currently selected folder
func (c *Client) FetchAllEmails(mbox *imap.MailboxStatus) ([]*Email, error) {
	if mbox.Messages == 0 {
		return nil, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mbox.Messages)

	return c.FetchEmails(seqSet)
}

// SearchByDate searches for emails older or newer than a date
func (c *Client) SearchByDate(before, after *time.Time) ([]uint32, error) {
	criteria := imap.NewSearchCriteria()

	if before != nil {
		criteria.Before = *before
	}
	if after != nil {
		criteria.Since = *after
	}

	uids, err := c.client.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return uids, nil
}

// DeleteMessages marks messages for deletion (expunge required to permanently delete)
func (c *Client) DeleteMessages(uids []uint32) error {
	if len(uids) == 0 {
		return nil
	}

	seqSet := new(imap.SeqSet)
	for _, uid := range uids {
		seqSet.AddNum(uid)
	}

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}

	if err := c.client.UidStore(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("failed to mark messages as deleted: %w", err)
	}

	return nil
}

// Expunge permanently removes messages marked for deletion
func (c *Client) Expunge() error {
	if err := c.client.Expunge(nil); err != nil {
		return fmt.Errorf("failed to expunge: %w", err)
	}
	return nil
}

// MoveMessages moves messages to another folder
func (c *Client) MoveMessages(uids []uint32, destFolder string) error {
	if len(uids) == 0 {
		return nil
	}

	seqSet := new(imap.SeqSet)
	for _, uid := range uids {
		seqSet.AddNum(uid)
	}

	// Copy to destination
	if err := c.client.UidCopy(seqSet, destFolder); err != nil {
		return fmt.Errorf("failed to copy messages to %s: %w", destFolder, err)
	}

	// Mark originals for deletion
	if err := c.DeleteMessages(uids); err != nil {
		return err
	}

	// Expunge to complete the move
	return c.Expunge()
}

// SetFlags sets flags on messages
func (c *Client) SetFlags(uids []uint32, flags []string, add bool) error {
	if len(uids) == 0 {
		return nil
	}

	seqSet := new(imap.SeqSet)
	for _, uid := range uids {
		seqSet.AddNum(uid)
	}

	var op imap.FlagsOp
	if add {
		op = imap.AddFlags
	} else {
		op = imap.RemoveFlags
	}

	item := imap.FormatFlagsOp(op, true)
	flagInterfaces := make([]interface{}, len(flags))
	for i, f := range flags {
		flagInterfaces[i] = f
	}

	if err := c.client.UidStore(seqSet, item, flagInterfaces, nil); err != nil {
		return fmt.Errorf("failed to set flags: %w", err)
	}

	return nil
}

// MarkAsRead marks messages as read (seen)
func (c *Client) MarkAsRead(uids []uint32) error {
	return c.SetFlags(uids, []string{imap.SeenFlag}, true)
}

// MarkAsUnread marks messages as unread (removes seen flag)
func (c *Client) MarkAsUnread(uids []uint32) error {
	return c.SetFlags(uids, []string{imap.SeenFlag}, false)
}

// FlagMessages adds the flagged flag to messages
func (c *Client) FlagMessages(uids []uint32) error {
	return c.SetFlags(uids, []string{imap.FlaggedFlag}, true)
}

// UnflagMessages removes the flagged flag from messages
func (c *Client) UnflagMessages(uids []uint32) error {
	return c.SetFlags(uids, []string{imap.FlaggedFlag}, false)
}

// hasAttachment checks if a body structure contains attachments
func hasAttachment(bs *imap.BodyStructure) bool {
	if bs == nil {
		return false
	}

	if bs.Disposition == "attachment" {
		return true
	}

	for _, part := range bs.Parts {
		if hasAttachment(part) {
			return true
		}
	}

	return false
}

// IsRead checks if an email has been read
func (e *Email) IsRead() bool {
	for _, flag := range e.Flags {
		if flag == imap.SeenFlag {
			return true
		}
	}
	return false
}

// IsFlagged checks if an email is flagged
func (e *Email) IsFlagged() bool {
	for _, flag := range e.Flags {
		if flag == imap.FlaggedFlag {
			return true
		}
	}
	return false
}

// LogInfo logs email info for debugging
func (e *Email) LogInfo() {
	log.Printf("  UID: %d, From: %s, Subject: %s, Date: %s, Read: %v",
		e.UID, e.From, e.Subject, e.Date.Format("2006-01-02"), e.IsRead())
}
