---
layout: default
title: Configuration Reference
---

# Configuration Reference

MailCleaner uses a JSON configuration file to define server settings and email processing rules.

## Configuration File Location

By default, MailCleaner looks for `config.json` in the current directory. You can specify a different location:

```bash
./mailcleaner -config /path/to/config.json
```

## Configuration Schema

```json
{
  "server": "imap.example.com",
  "port": 993,
  "username": "your-email@example.com",
  "password": "your-password",
  "tls": true,
  "rules": [
    {
      "sender": "pattern",
      "move_to_folder": "FolderName"
    }
  ]
}
```

## Server Settings

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `server` | string | Yes | - | IMAP server hostname |
| `port` | integer | Yes | - | IMAP server port (usually 993 for TLS, 143 for plaintext) |
| `username` | string | Yes | - | Email account username |
| `password` | string | Yes | - | Email account password or app password |
| `tls` | boolean | No | `true` | Enable TLS encryption |

## Rules

Rules define which emails to match and where to move them. Rules are processed in order.

### Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `sender` | string | Yes | Pattern to match against the sender's email address |
| `move_to_folder` | string | Yes | Destination folder for matched emails |

### Pattern Matching

The `sender` field performs **case-insensitive partial matching** on the email address:

| Pattern | Matches | Does Not Match |
|---------|---------|----------------|
| `newsletter@` | `newsletter@company.com`, `newsletter@example.org` | `my-newsletter@test.com` |
| `@github.com` | `noreply@github.com`, `notifications@github.com` | `github@example.com` |
| `john` | `john@example.com`, `johnny@test.com`, `user@john.com` | `Jane@example.com` |
| `support@acme.com` | `support@acme.com` | `support@acme.org` |

## Example Configurations

### Newsletter Organization

```json
{
  "server": "imap.gmail.com",
  "port": 993,
  "username": "user@gmail.com",
  "password": "your-app-password",
  "rules": [
    {
      "sender": "newsletter@",
      "move_to_folder": "Newsletters"
    },
    {
      "sender": "digest@",
      "move_to_folder": "Newsletters"
    },
    {
      "sender": "weekly@",
      "move_to_folder": "Newsletters"
    }
  ]
}
```

### GitHub Notifications

```json
{
  "server": "imap.gmail.com",
  "port": 993,
  "username": "user@gmail.com",
  "password": "your-app-password",
  "rules": [
    {
      "sender": "notifications@github.com",
      "move_to_folder": "GitHub/Notifications"
    },
    {
      "sender": "noreply@github.com",
      "move_to_folder": "GitHub/General"
    }
  ]
}
```

### Multi-Service Organization

```json
{
  "server": "imap.example.com",
  "port": 993,
  "username": "user@example.com",
  "password": "your-password",
  "rules": [
    {
      "sender": "@github.com",
      "move_to_folder": "Services/GitHub"
    },
    {
      "sender": "@gitlab.com",
      "move_to_folder": "Services/GitLab"
    },
    {
      "sender": "@slack.com",
      "move_to_folder": "Services/Slack"
    },
    {
      "sender": "@trello.com",
      "move_to_folder": "Services/Trello"
    }
  ]
}
```

## Security Considerations

- **Never commit** your `config.json` to version control
- Use **app-specific passwords** when available
- Consider using **environment variables** or a secrets manager in production
- The `config.json` file should have restricted permissions (`chmod 600 config.json`)

## Next Steps

- [Usage Examples](usage) - See MailCleaner in action
- [Getting Started](getting-started) - Return to setup guide
