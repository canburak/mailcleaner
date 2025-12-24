package rules

import (
	"testing"
	"time"

	"github.com/mailcleaner/mailcleaner/internal/config"
	"github.com/mailcleaner/mailcleaner/internal/imap"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestMatchConditions(t *testing.T) {
	engine := NewEngine(&config.Config{}, true)

	tests := []struct {
		name       string
		email      *imap.Email
		conditions config.Conditions
		expected   bool
	}{
		{
			name: "from exact match",
			email: &imap.Email{
				From: "test@example.com",
			},
			conditions: config.Conditions{
				From: "test@example.com",
			},
			expected: true,
		},
		{
			name: "from exact match case insensitive",
			email: &imap.Email{
				From: "Test@Example.com",
			},
			conditions: config.Conditions{
				From: "test@example.com",
			},
			expected: true,
		},
		{
			name: "from exact no match",
			email: &imap.Email{
				From: "other@example.com",
			},
			conditions: config.Conditions{
				From: "test@example.com",
			},
			expected: false,
		},
		{
			name: "from contains match",
			email: &imap.Email{
				From: "newsletter@company.com",
			},
			conditions: config.Conditions{
				FromContains: "@company",
			},
			expected: true,
		},
		{
			name: "from contains no match",
			email: &imap.Email{
				From: "newsletter@other.com",
			},
			conditions: config.Conditions{
				FromContains: "@company",
			},
			expected: false,
		},
		{
			name: "subject contains match",
			email: &imap.Email{
				Subject: "Weekly Newsletter: Important Updates",
			},
			conditions: config.Conditions{
				SubjectContains: "newsletter",
			},
			expected: true,
		},
		{
			name: "subject contains case insensitive",
			email: &imap.Email{
				Subject: "URGENT: Action Required",
			},
			conditions: config.Conditions{
				SubjectContains: "urgent",
			},
			expected: true,
		},
		{
			name: "older than days match",
			email: &imap.Email{
				Date: time.Now().AddDate(0, 0, -35),
			},
			conditions: config.Conditions{
				OlderThanDays: 30,
			},
			expected: true,
		},
		{
			name: "older than days no match",
			email: &imap.Email{
				Date: time.Now().AddDate(0, 0, -10),
			},
			conditions: config.Conditions{
				OlderThanDays: 30,
			},
			expected: false,
		},
		{
			name: "newer than days match",
			email: &imap.Email{
				Date: time.Now().AddDate(0, 0, -5),
			},
			conditions: config.Conditions{
				NewerThanDays: 10,
			},
			expected: true,
		},
		{
			name: "newer than days no match",
			email: &imap.Email{
				Date: time.Now().AddDate(0, 0, -15),
			},
			conditions: config.Conditions{
				NewerThanDays: 10,
			},
			expected: false,
		},
		{
			name: "is read true match",
			email: &imap.Email{
				Flags: []string{"\\Seen"},
			},
			conditions: config.Conditions{
				IsRead: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "is read true no match",
			email: &imap.Email{
				Flags: []string{},
			},
			conditions: config.Conditions{
				IsRead: boolPtr(true),
			},
			expected: false,
		},
		{
			name: "is unread true match",
			email: &imap.Email{
				Flags: []string{},
			},
			conditions: config.Conditions{
				IsUnread: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "has attachment match",
			email: &imap.Email{
				HasAttachment: true,
			},
			conditions: config.Conditions{
				HasAttachment: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "has attachment no match",
			email: &imap.Email{
				HasAttachment: false,
			},
			conditions: config.Conditions{
				HasAttachment: boolPtr(true),
			},
			expected: false,
		},
		{
			name: "size larger than match",
			email: &imap.Email{
				Size: 15 * 1024 * 1024, // 15MB
			},
			conditions: config.Conditions{
				SizeLargerThan: "10MB",
			},
			expected: true,
		},
		{
			name: "size larger than no match",
			email: &imap.Email{
				Size: 5 * 1024 * 1024, // 5MB
			},
			conditions: config.Conditions{
				SizeLargerThan: "10MB",
			},
			expected: false,
		},
		{
			name: "multiple conditions all match",
			email: &imap.Email{
				From:    "newsletter@company.com",
				Subject: "Weekly Update",
				Date:    time.Now().AddDate(0, 0, -45),
				Flags:   []string{"\\Seen"},
			},
			conditions: config.Conditions{
				FromContains:    "@company",
				SubjectContains: "update",
				OlderThanDays:   30,
				IsRead:          boolPtr(true),
			},
			expected: true,
		},
		{
			name: "multiple conditions one fails",
			email: &imap.Email{
				From:    "newsletter@company.com",
				Subject: "Weekly Update",
				Date:    time.Now().AddDate(0, 0, -45),
				Flags:   []string{}, // Not read
			},
			conditions: config.Conditions{
				FromContains:    "@company",
				SubjectContains: "update",
				OlderThanDays:   30,
				IsRead:          boolPtr(true),
			},
			expected: false,
		},
		{
			name: "to contains match",
			email: &imap.Email{
				To: []string{"support@company.com", "admin@company.com"},
			},
			conditions: config.Conditions{
				ToContains: "admin",
			},
			expected: true,
		},
		{
			name: "empty conditions matches all",
			email: &imap.Email{
				From:    "anyone@anywhere.com",
				Subject: "Anything",
			},
			conditions: config.Conditions{},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.matchConditions(tt.email, &tt.conditions)
			if result != tt.expected {
				t.Errorf("matchConditions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected uint32
		hasError bool
	}{
		{"10MB", 10 * 1024 * 1024, false},
		{"10mb", 10 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"500KB", 500 * 1024, false},
		{"1024B", 1024, false},
		{"100", 100, false},
		{" 50 MB ", 50 * 1024 * 1024, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("parseSize(%s) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestMatchEmails(t *testing.T) {
	engine := NewEngine(&config.Config{}, true)

	emails := []*imap.Email{
		{UID: 1, From: "newsletter@company.com", Date: time.Now().AddDate(0, 0, -45)},
		{UID: 2, From: "friend@personal.com", Date: time.Now().AddDate(0, 0, -10)},
		{UID: 3, From: "promo@company.com", Date: time.Now().AddDate(0, 0, -35)},
		{UID: 4, From: "newsletter@other.com", Date: time.Now().AddDate(0, 0, -50)},
	}

	conditions := config.Conditions{
		FromContains:  "@company",
		OlderThanDays: 30,
	}

	matched := engine.matchEmails(emails, &conditions)

	if len(matched) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matched))
	}

	expectedUIDs := map[uint32]bool{1: true, 3: true}
	for _, email := range matched {
		if !expectedUIDs[email.UID] {
			t.Errorf("unexpected matched email UID: %d", email.UID)
		}
	}
}
