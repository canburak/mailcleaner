---
layout: default
title: Configuration Reference
---

# Configuration Reference

MailCleaner supports two configuration methods: the web UI for interactive use, and JSON files for CLI automation.

## Web UI Configuration

### Accounts

Accounts are managed through the web interface at `/accounts`. Each account requires:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name for the account |
| `server` | string | Yes | IMAP server hostname |
| `port` | integer | Yes | IMAP server port (usually 993) |
| `username` | string | Yes | Email account username |
| `password` | string | Yes | Email account password |
| `tls` | boolean | No | Enable TLS (default: true) |

### Rules

Rules are created per account through the web interface. Each rule has:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name for the rule |
| `pattern` | string | Yes | Pattern to match |
| `pattern_type` | string | Yes | Type of matching (see below) |
| `move_to_folder` | string | Yes | Destination folder |
| `enabled` | boolean | No | Whether rule is active (default: true) |
| `priority` | integer | No | Rule priority (lower = higher priority) |

### Pattern Types

| Type | Description | Example Pattern | Matches |
|------|-------------|-----------------|---------|
| `sender` | Match the From address | `newsletter@` | `newsletter@company.com` |
| `subject` | Match the subject line | `[URGENT]` | Subjects containing `[URGENT]` |
| `from_domain` | Match sender's domain | `github.com` | All emails from `@github.com` |

All patterns are **case-insensitive partial matches**.

### Web UI Rule Example

```json
{
  "name": "GitHub Notifications",
  "pattern": "github.com",
  "pattern_type": "from_domain",
  "move_to_folder": "GitHub",
  "enabled": true,
  "priority": 10
}
```

## CLI Configuration

The CLI uses a JSON configuration file for backwards compatibility.

### Configuration File Location

By default, the CLI looks for `config.json` in the current directory:

```bash
./mailcleaner -config /path/to/config.json
```

### CLI Configuration Schema

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

### CLI Configuration Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `server` | string | Yes | - | IMAP server hostname |
| `port` | integer | Yes | - | IMAP server port |
| `username` | string | Yes | - | Email account username |
| `password` | string | Yes | - | Email account password |
| `tls` | boolean | No | `true` | Enable TLS encryption |
| `rules` | array | Yes | - | Array of rule objects |

### CLI Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `sender` | string | Yes | Pattern to match against sender |
| `move_to_folder` | string | Yes | Destination folder |

## Example Configurations

### Newsletter Organization (Web UI)

Create these rules in the Rules section:

| Name | Pattern | Type | Folder |
|------|---------|------|--------|
| Newsletters | `newsletter@` | sender | Newsletters |
| Digests | `digest@` | sender | Newsletters |
| Weekly Updates | `weekly@` | sender | Newsletters |

### GitHub Notifications (Web UI)

```json
[
  {
    "name": "GitHub Notifications",
    "pattern": "notifications@github.com",
    "pattern_type": "sender",
    "move_to_folder": "GitHub/Notifications",
    "priority": 1
  },
  {
    "name": "GitHub General",
    "pattern": "github.com",
    "pattern_type": "from_domain",
    "move_to_folder": "GitHub/General",
    "priority": 10
  }
]
```

### Multi-Service Organization (CLI)

```json
{
  "server": "imap.example.com",
  "port": 993,
  "username": "user@example.com",
  "password": "your-password",
  "rules": [
    { "sender": "@github.com", "move_to_folder": "Services/GitHub" },
    { "sender": "@gitlab.com", "move_to_folder": "Services/GitLab" },
    { "sender": "@slack.com", "move_to_folder": "Services/Slack" },
    { "sender": "@trello.com", "move_to_folder": "Services/Trello" }
  ]
}
```

## Database

The web server stores data in SQLite at `~/.mailcleaner/data.db` by default.

- Schema is automatically migrated on startup
- Use `-db` flag to specify a different location
- Passwords are stored in the database (consider security implications)

## Security Considerations

- **Never commit** credentials to version control
- Use **app-specific passwords** when available
- Run behind a **reverse proxy** in production
- Consider **OAuth2/XOAUTH2** for Gmail and other providers
- Restrict file permissions: `chmod 600 config.json`

## Next Steps

- [Usage Examples](usage) - See MailCleaner in action
- [API Reference](api) - REST API documentation
- [Getting Started](getting-started) - Return to setup guide
