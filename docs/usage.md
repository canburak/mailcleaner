---
layout: default
title: Usage Examples
---

# Usage Examples

This page covers common use cases for both the web UI and CLI.

## Web UI Usage

### Managing Accounts

#### Adding a New Account

1. Navigate to **Accounts** in the sidebar
2. Click **Add Account**
3. Fill in the connection details:
   - **Name**: A friendly name (e.g., "Work Email")
   - **Server**: IMAP server hostname
   - **Port**: Usually 993 for TLS
   - **Username**: Your email address
   - **Password**: Your password or app password
4. Click **Test Connection** to verify
5. Click **Save**

#### Testing Connection

After saving an account, use the **Test Connection** button to:
- Verify credentials are correct
- See available IMAP folders
- Check the total email count

### Creating Rules

#### Basic Rule Creation

1. Select an account from **Accounts**
2. Navigate to **Rules**
3. Click **Add Rule**
4. Configure the rule:
   - **Name**: Descriptive name
   - **Pattern**: Text to match
   - **Pattern Type**: sender, subject, or from_domain
   - **Move to Folder**: Destination folder
   - **Priority**: Lower numbers run first
5. Click **Save**

#### Rule Priority

Rules are applied in priority order (lowest number first). Use this to:
- Apply specific rules before general ones
- Create catch-all rules with high priority numbers

Example priority ordering:
```
Priority 1:  Match notifications@github.com → GitHub/Notifications
Priority 10: Match @github.com → GitHub/General
Priority 100: Match newsletter@ → Newsletters
```

### Preview and Apply

#### Live Preview

1. Go to **Preview** for an account
2. Select the source folder (usually INBOX)
3. View emails and which rules match them
4. Matched emails show the destination folder

The preview uses WebSocket for real-time updates as you modify rules.

#### Applying Rules

1. Review the preview carefully
2. Click **Apply Rules**
3. Confirm the action
4. Watch the progress as emails are moved

**Warning**: Moving emails is not easily reversible. Always preview first!

## CLI Usage

### Command-Line Options

```bash
./mailcleaner [options]
```

| Option | Description |
|--------|-------------|
| `-config <path>` | Path to configuration file (default: `config.json`) |
| `-dry-run` | Preview changes without moving emails |

### Dry Run (Recommended First Step)

Always test your configuration first:

```bash
./mailcleaner -dry-run
```

Example output:

```
Connecting to imap.gmail.com:993...
Connected successfully
Processing INBOX (1523 messages)
[DRY RUN] Would move: "Weekly Newsletter" from newsletter@example.com -> Newsletters
[DRY RUN] Would move: "Your PR was merged" from notifications@github.com -> GitHub
Dry run complete. 2 emails would be moved.
```

### Production Run

Once satisfied with the dry run:

```bash
./mailcleaner
```

### Custom Config File

```bash
./mailcleaner -config ~/configs/work-email.json
```

## Common Use Cases

### Organizing a Cluttered Inbox (Web UI)

Create these rules with ascending priority:

| Priority | Name | Pattern | Type | Folder |
|----------|------|---------|------|--------|
| 1 | GitHub PRs | `[Pull Request]` | subject | GitHub/PRs |
| 2 | GitHub Issues | `[Issue]` | subject | GitHub/Issues |
| 10 | GitHub All | `github.com` | from_domain | GitHub/Other |
| 20 | Newsletters | `newsletter@` | sender | Newsletters |
| 21 | Digests | `digest@` | sender | Newsletters |

### Separating Work and Personal (CLI)

```json
{
  "server": "imap.company.com",
  "port": 993,
  "username": "user@company.com",
  "password": "app-password",
  "rules": [
    { "sender": "@company.slack.com", "move_to_folder": "Work/Slack" },
    { "sender": "@jira.atlassian.com", "move_to_folder": "Work/Jira" },
    { "sender": "@github.com", "move_to_folder": "Work/GitHub" }
  ]
}
```

### Scheduled Automation (CLI)

#### Linux/macOS (cron)

```cron
# Run every hour
0 * * * * /path/to/mailcleaner -config /path/to/config.json >> /var/log/mailcleaner.log 2>&1
```

#### systemd Timer

Create `/etc/systemd/system/mailcleaner.service`:

```ini
[Unit]
Description=MailCleaner Email Organizer

[Service]
Type=oneshot
ExecStart=/path/to/mailcleaner -config /path/to/config.json
User=youruser
```

Create `/etc/systemd/system/mailcleaner.timer`:

```ini
[Unit]
Description=Run MailCleaner hourly

[Timer]
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target
```

Enable with: `systemctl enable --now mailcleaner.timer`

## Tips and Best Practices

1. **Always preview first** - Use dry-run (CLI) or Preview (Web UI)
2. **Be specific with patterns** - More specific patterns reduce false matches
3. **Create folders first** - Destination folders must exist
4. **Order matters** - Put specific rules before general ones
5. **Use priority** - In Web UI, lower priority numbers run first
6. **Back up important emails** - Before large operations

## Troubleshooting

### No Emails Matched

- Check pattern spelling and case
- Verify pattern type is correct
- Try a more general pattern to test

### Authentication Failed

- Verify username and password
- Check if 2FA requires an app password
- Ensure IMAP is enabled in email settings
- For Gmail, create an [App Password](https://support.google.com/accounts/answer/185833)

### Folder Not Found

- Create the folder in your email client first
- Check folder name capitalization
- Use the correct path separator (usually `/`)

### WebSocket Connection Failed

- Check that the server is running
- Verify no firewall is blocking WebSocket
- Try refreshing the browser

## Next Steps

- [Configuration Reference](configuration) - Full configuration options
- [API Reference](api) - REST API documentation
- [Getting Started](getting-started) - Installation guide
