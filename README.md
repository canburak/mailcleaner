# MailCleaner

A simple email organizer that matches emails by sender and moves them to folders.

## Features

- Connect to any IMAP server
- Match emails by sender address (partial match)
- Move matched emails to specified folders
- Dry-run mode to preview changes

## Installation

```bash
go build -o mailcleaner
```

## Configuration

Create a `config.json` file (see `config.example.json`):

```json
{
  "server": "imap.example.com",
  "port": 993,
  "username": "your-email@example.com",
  "password": "your-password",
  "rules": [
    {
      "sender": "newsletter@",
      "move_to_folder": "Newsletters"
    }
  ]
}
```

### Rule Matching

The `sender` field performs a case-insensitive partial match on the email address:
- `"newsletter@"` matches `newsletter@company.com`
- `"@github.com"` matches any email from GitHub
- `"john"` matches `john@example.com` or `johnny@test.com`

## Usage

```bash
# Preview what would be moved (dry run)
./mailcleaner -dry-run

# Run and move matching emails
./mailcleaner

# Use custom config file
./mailcleaner -config /path/to/config.json
```

## Supported Protocols

- **IMAP** - Internet Message Access Protocol (TLS enabled by default, set `"tls": false` for plaintext)
