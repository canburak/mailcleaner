// Package testserver provides an in-memory IMAP server for testing
package testserver

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/emersion/go-imap/server"
)

// TestServer is an in-memory IMAP server for testing
type TestServer struct {
	server   *server.Server
	listener net.Listener
	backend  *MemoryBackend
	Addr     string
}

// New creates a new test IMAP server
func New(user, pass string) (*TestServer, error) {
	be := NewMemoryBackend(user, pass)

	s := server.New(be)
	s.AllowInsecureAuth = true

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	ts := &TestServer{
		server:   s,
		listener: listener,
		backend:  be,
		Addr:     listener.Addr().String(),
	}

	go s.Serve(listener)

	return ts, nil
}

// Close shuts down the test server
func (ts *TestServer) Close() error {
	return ts.listener.Close()
}

// AddMessage adds a test message to the user's INBOX
func (ts *TestServer) AddMessage(from, subject, body string) {
	ts.backend.AddMessage("INBOX", from, subject, body)
}

// AddMessageToFolder adds a test message to a specific folder
func (ts *TestServer) AddMessageToFolder(folder, from, subject, body string) {
	ts.backend.AddMessage(folder, from, subject, body)
}

// GetMessageCount returns the number of messages in a folder
func (ts *TestServer) GetMessageCount(folder string) int {
	return ts.backend.GetMessageCount(folder)
}

// CreateFolder creates a new mailbox folder
func (ts *TestServer) CreateFolder(name string) {
	ts.backend.CreateMailbox(name)
}

// MemoryBackend is an in-memory IMAP backend
type MemoryBackend struct {
	user     *MemoryUser
	username string
	password string
}

// NewMemoryBackend creates a new memory backend
func NewMemoryBackend(username, password string) *MemoryBackend {
	be := &MemoryBackend{
		username: username,
		password: password,
	}
	be.user = &MemoryUser{
		username:  username,
		password:  password,
		mailboxes: make(map[string]*MemoryMailbox),
	}
	// Create default INBOX
	be.user.mailboxes["INBOX"] = &MemoryMailbox{
		name:     "INBOX",
		messages: []*MemoryMessage{},
		uidNext:  1,
	}
	return be
}

func (be *MemoryBackend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	if username != be.username || password != be.password {
		return nil, errors.New("invalid credentials")
	}
	return be.user, nil
}

func (be *MemoryBackend) AddMessage(folder, from, subject, body string) {
	be.user.mu.Lock()
	defer be.user.mu.Unlock()

	mbox, ok := be.user.mailboxes[folder]
	if !ok {
		mbox = &MemoryMailbox{
			name:     folder,
			messages: []*MemoryMessage{},
			uidNext:  1,
		}
		be.user.mailboxes[folder] = mbox
	}

	msg := &MemoryMessage{
		uid:     mbox.uidNext,
		from:    from,
		subject: subject,
		body:    body,
		date:    time.Now(),
		flags:   []string{},
	}
	mbox.messages = append(mbox.messages, msg)
	mbox.uidNext++
}

func (be *MemoryBackend) GetMessageCount(folder string) int {
	be.user.mu.RLock()
	defer be.user.mu.RUnlock()

	mbox, ok := be.user.mailboxes[folder]
	if !ok {
		return 0
	}
	count := 0
	for _, m := range mbox.messages {
		if !m.deleted {
			count++
		}
	}
	return count
}

func (be *MemoryBackend) CreateMailbox(name string) {
	be.user.mu.Lock()
	defer be.user.mu.Unlock()

	if _, ok := be.user.mailboxes[name]; !ok {
		be.user.mailboxes[name] = &MemoryMailbox{
			name:     name,
			messages: []*MemoryMessage{},
			uidNext:  1,
		}
	}
}

// MemoryUser represents an in-memory user
type MemoryUser struct {
	username  string
	password  string
	mailboxes map[string]*MemoryMailbox
	mu        sync.RWMutex
}

func (u *MemoryUser) Username() string {
	return u.username
}

func (u *MemoryUser) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var mailboxes []backend.Mailbox
	for _, mbox := range u.mailboxes {
		mailboxes = append(mailboxes, mbox)
	}
	return mailboxes, nil
}

func (u *MemoryUser) GetMailbox(name string) (backend.Mailbox, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	mbox, ok := u.mailboxes[name]
	if !ok {
		return nil, errors.New("mailbox not found")
	}
	return mbox, nil
}

func (u *MemoryUser) CreateMailbox(name string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.mailboxes[name]; ok {
		return errors.New("mailbox already exists")
	}
	u.mailboxes[name] = &MemoryMailbox{
		name:     name,
		messages: []*MemoryMessage{},
		uidNext:  1,
	}
	return nil
}

func (u *MemoryUser) DeleteMailbox(name string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if name == "INBOX" {
		return errors.New("cannot delete INBOX")
	}
	delete(u.mailboxes, name)
	return nil
}

func (u *MemoryUser) RenameMailbox(existingName, newName string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	mbox, ok := u.mailboxes[existingName]
	if !ok {
		return errors.New("mailbox not found")
	}
	mbox.name = newName
	u.mailboxes[newName] = mbox
	delete(u.mailboxes, existingName)
	return nil
}

func (u *MemoryUser) Logout() error {
	return nil
}

// MemoryMailbox represents an in-memory mailbox
type MemoryMailbox struct {
	name     string
	messages []*MemoryMessage
	uidNext  uint32
	mu       sync.RWMutex
}

func (m *MemoryMailbox) Name() string {
	return m.name
}

func (m *MemoryMailbox) Info() (*imap.MailboxInfo, error) {
	return &imap.MailboxInfo{
		Name:       m.name,
		Delimiter:  "/",
		Attributes: nil,
	}, nil
}

func (m *MemoryMailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := imap.NewMailboxStatus(m.name, items)
	status.Messages = uint32(len(m.messages))
	status.UidNext = m.uidNext
	status.UidValidity = 1
	return status, nil
}

func (m *MemoryMailbox) SetSubscribed(subscribed bool) error {
	return nil
}

func (m *MemoryMailbox) Check() error {
	return nil
}

func (m *MemoryMailbox) ListMessages(uid bool, seqSet *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	defer close(ch)
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i, msg := range m.messages {
		if msg.deleted {
			continue
		}

		seqNum := uint32(i + 1)
		var match bool
		if uid {
			match = seqSet.Contains(msg.uid)
		} else {
			match = seqSet.Contains(seqNum)
		}

		if match {
			ch <- msg.ToIMAP(seqNum, items)
		}
	}
	return nil
}

func (m *MemoryMailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []uint32
	for i, msg := range m.messages {
		if msg.deleted {
			continue
		}
		if uid {
			results = append(results, msg.uid)
		} else {
			results = append(results, uint32(i+1))
		}
	}
	return results, nil
}

func (m *MemoryMailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := &MemoryMessage{
		uid:   m.uidNext,
		date:  date,
		flags: flags,
	}
	m.messages = append(m.messages, msg)
	m.uidNext++
	return nil
}

func (m *MemoryMailbox) UpdateMessagesFlags(uid bool, seqSet *imap.SeqSet, op imap.FlagsOp, flags []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, msg := range m.messages {
		seqNum := uint32(i + 1)
		var match bool
		if uid {
			match = seqSet.Contains(msg.uid)
		} else {
			match = seqSet.Contains(seqNum)
		}

		if match {
			switch op {
			case imap.AddFlags:
				for _, f := range flags {
					if f == imap.DeletedFlag {
						msg.deleted = true
					}
					msg.flags = append(msg.flags, f)
				}
			case imap.RemoveFlags:
				for _, f := range flags {
					if f == imap.DeletedFlag {
						msg.deleted = false
					}
				}
			case imap.SetFlags:
				msg.flags = flags
				msg.deleted = false
				for _, f := range flags {
					if f == imap.DeletedFlag {
						msg.deleted = true
					}
				}
			}
		}
	}
	return nil
}

func (m *MemoryMailbox) CopyMessages(uid bool, seqSet *imap.SeqSet, destName string) error {
	// This is handled at the user level
	return errors.New("copy not implemented at mailbox level")
}

func (m *MemoryMailbox) Expunge() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var remaining []*MemoryMessage
	for _, msg := range m.messages {
		if !msg.deleted {
			remaining = append(remaining, msg)
		}
	}
	m.messages = remaining
	return nil
}

// MemoryMessage represents an in-memory message
type MemoryMessage struct {
	uid     uint32
	from    string
	subject string
	body    string
	date    time.Time
	flags   []string
	deleted bool
}

func (m *MemoryMessage) ToIMAP(seqNum uint32, items []imap.FetchItem) *imap.Message {
	msg := imap.NewMessage(seqNum, items)
	msg.Uid = m.uid

	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			msg.Envelope = &imap.Envelope{
				Subject: m.subject,
				From:    parseAddress(m.from),
				Date:    m.date,
			}
		case imap.FetchFlags:
			msg.Flags = m.flags
		case imap.FetchUid:
			msg.Uid = m.uid
		}
	}
	return msg
}

func parseAddress(email string) []*imap.Address {
	if email == "" {
		return nil
	}
	// Simple parsing: split on @
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return []*imap.Address{{
				MailboxName: email[:i],
				HostName:    email[i+1:],
			}}
		}
	}
	return []*imap.Address{{MailboxName: email}}
}
