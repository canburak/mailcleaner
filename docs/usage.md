---
layout: default
title: Usage Examples
---

# Usage Examples

This page covers common use cases and command-line options for MailCleaner.

## Command-Line Options

```bash
./mailcleaner [options]
```

| Option | Description |
|--------|-------------|
| `-config <path>` | Path to configuration file (default: `config.json`) |
| `-dry-run` | Preview changes without moving emails |

## Common Workflows

### Preview Mode (Recommended First Step)

Always run in dry-run mode first to see what would be moved:

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
[DRY RUN] Would move: "Security Alert" from noreply@github.com -> GitHub
Dry run complete. 3 emails would be moved.
```

### Production Run

Once satisfied with the dry run output:

```bash
./mailcleaner
```

### Custom Configuration File

Use a different configuration file:

```bash
./mailcleaner -config ~/configs/work-email.json
```

## Use Case Examples

### Organizing a Cluttered Inbox

**Problem**: Your inbox has thousands of unread newsletters and notifications.

**Solution**: Create rules for common senders:

```json
{
  "rules": [
    { "sender": "newsletter@", "move_to_folder": "Newsletters" },
    { "sender": "digest@", "move_to_folder": "Newsletters" },
    { "sender": "noreply@", "move_to_folder": "Automated" },
    { "sender": "notifications@", "move_to_folder": "Notifications" }
  ]
}
```

### Separating Work and Personal

**Problem**: Work notifications mixed with personal emails.

**Solution**: Create service-specific folders:

```json
{
  "rules": [
    { "sender": "@company.slack.com", "move_to_folder": "Work/Slack" },
    { "sender": "@jira.atlassian.com", "move_to_folder": "Work/Jira" },
    { "sender": "@github.com", "move_to_folder": "Work/GitHub" },
    { "sender": "@confluence.atlassian.com", "move_to_folder": "Work/Confluence" }
  ]
}
```

### Archiving Old Subscriptions

**Problem**: You want to archive emails from services you no longer actively use.

**Solution**: Move them to an archive folder:

```json
{
  "rules": [
    { "sender": "@oldservice.com", "move_to_folder": "Archive/OldService" },
    { "sender": "@deprecatedapp.io", "move_to_folder": "Archive/Deprecated" }
  ]
}
```

## Running as a Scheduled Task

### Linux/macOS (cron)

Add to your crontab (`crontab -e`):

```cron
# Run every hour
0 * * * * /path/to/mailcleaner -config /path/to/config.json >> /var/log/mailcleaner.log 2>&1
```

### macOS (launchd)

Create `~/Library/LaunchAgents/com.mailcleaner.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.mailcleaner</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/mailcleaner</string>
        <string>-config</string>
        <string>/path/to/config.json</string>
    </array>
    <key>StartInterval</key>
    <integer>3600</integer>
</dict>
</plist>
```

Load with: `launchctl load ~/Library/LaunchAgents/com.mailcleaner.plist`

## Tips and Best Practices

1. **Always dry-run first** - Verify rules before moving emails
2. **Be specific with patterns** - More specific patterns reduce false matches
3. **Create folders first** - Ensure destination folders exist before running
4. **Order matters** - Rules are processed in order; put specific rules before general ones
5. **Back up important emails** - Consider backing up before large operations

## Troubleshooting

### No Emails Matched

- Check your sender patterns are correct
- Verify the INBOX has emails matching your criteria
- Use more general patterns to test (e.g., `@` matches all emails)

### Authentication Failed

- Verify username and password
- Check if 2FA requires an app password
- Ensure IMAP is enabled in your email settings

### Folder Not Found

- Create the destination folder in your email client first
- Check folder name capitalization
- Use full path for nested folders (e.g., `INBOX/Newsletters`)

## Next Steps

- [Configuration Reference](configuration) - Full configuration options
- [Getting Started](getting-started) - Installation guide
