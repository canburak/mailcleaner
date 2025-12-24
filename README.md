# MailCleaner

An application designed to run rules periodically on a remote mailbox.

## Overview

MailCleaner connects to your email mailbox and executes user-defined rules on a schedule. This allows you to automate email organization, cleanup, and management tasks without manual intervention.

## Features

- **IMAP Support** - Connect to any IMAP-compatible email server
- **Flexible Rules** - Match emails by sender, subject, date, size, flags, and more
- **Multiple Actions** - Delete, move, mark read/unread, flag emails
- **Scheduled Execution** - Run rules on an interval or cron schedule
- **Dry Run Mode** - Preview changes without modifying your mailbox
- **Multiple Accounts** - Manage rules across multiple email accounts

## Installation

```bash
# Clone the repository
git clone https://github.com/mailcleaner/mailcleaner.git
cd mailcleaner

# Build
go build -o mailcleaner ./cmd/mailcleaner

# Or install directly
go install ./cmd/mailcleaner
```

## Quick Start

1. Copy the example configuration:
   ```bash
   cp config.example.yaml config.yaml
   ```

2. Edit `config.yaml` with your email account and rules

3. Set your password as an environment variable:
   ```bash
   export MAILCLEANER_GMAIL_PASSWORD="your-app-password"
   ```

4. Validate your configuration:
   ```bash
   ./mailcleaner validate
   ```

5. Run once in dry-run mode to preview:
   ```bash
   ./mailcleaner run --dry-run
   ```

6. Run for real:
   ```bash
   ./mailcleaner run
   ```

## Usage

```bash
# Run all rules once
mailcleaner run

# Run in dry-run mode (no changes made)
mailcleaner run --dry-run

# Run as a daemon with scheduled execution
mailcleaner daemon

# Validate configuration file
mailcleaner validate

# List folders in a mailbox
mailcleaner list-folders
mailcleaner list-folders --account work

# Use a different config file
mailcleaner run --config /path/to/config.yaml

# Show version
mailcleaner version
```

## Configuration

See `config.example.yaml` for a complete example. Here's a quick overview:

### Accounts

```yaml
accounts:
  - name: personal
    host: imap.gmail.com
    port: 993
    tls: true
    username: your.email@gmail.com
    password_env: MAILCLEANER_GMAIL_PASSWORD  # Read from environment
```

### Rules

```yaml
rules:
  - name: Delete old newsletters
    account: personal
    folder: INBOX
    conditions:
      from_contains: "@newsletter"
      older_than_days: 30
    action:
      type: delete
```

### Available Conditions

| Condition | Description |
|-----------|-------------|
| `from` | Exact match on sender email |
| `from_contains` | Sender email contains string |
| `to` | Exact match on recipient |
| `to_contains` | Recipient contains string |
| `subject` | Exact match on subject |
| `subject_contains` | Subject contains string |
| `older_than_days` | Email is older than N days |
| `newer_than_days` | Email is newer than N days |
| `is_read` | Email has been read (true/false) |
| `is_unread` | Email has not been read (true/false) |
| `has_attachment` | Email has attachments (true/false) |
| `size_larger_than` | Email size exceeds (e.g., "10MB", "500KB") |

### Available Actions

| Action | Parameters | Description |
|--------|------------|-------------|
| `delete` | - | Delete matched emails |
| `move` | `move_to: "Folder"` | Move to specified folder |
| `mark_read` | - | Mark as read |
| `mark_unread` | - | Mark as unread |
| `flag` | - | Add starred/flagged status |
| `unflag` | - | Remove starred/flagged status |
| `add_flag` | `add_flag: "CustomFlag"` | Add custom flag |
| `remove_flag` | `remove_flag: "CustomFlag"` | Remove custom flag |

### Schedule

```yaml
schedule:
  # Run every N minutes
  interval_minutes: 60

  # Or use cron expression (overrides interval)
  cron: "0 6 * * *"  # Daily at 6 AM
```

## Gmail Setup

For Gmail, you need to use an App Password:

1. Enable 2-Factor Authentication on your Google account
2. Go to [Google App Passwords](https://myaccount.google.com/apppasswords)
3. Generate a new app password for "Mail"
4. Use this password in your configuration

Common Gmail IMAP folders:
- `INBOX`
- `[Gmail]/All Mail`
- `[Gmail]/Trash`
- `[Gmail]/Spam`
- `[Gmail]/Drafts`
- `[Gmail]/Sent Mail`
- `[Gmail]/Starred`
- `[Gmail]/Important`
- `[Gmail]/Promotions`
- `[Gmail]/Social`
- `[Gmail]/Updates`

## Running as a Service

### systemd (Linux)

Create `/etc/systemd/system/mailcleaner.service`:

```ini
[Unit]
Description=MailCleaner Email Automation
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/mailcleaner
ExecStart=/path/to/mailcleaner daemon --config /path/to/config.yaml
Restart=always
RestartSec=10
EnvironmentFile=/path/to/.env

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl enable mailcleaner
sudo systemctl start mailcleaner
```

## Development

```bash
# Run tests
go test ./...

# Build
go build -o mailcleaner ./cmd/mailcleaner

# Run with race detector
go run -race ./cmd/mailcleaner run --dry-run
```

## License

MIT
