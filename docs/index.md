---
layout: default
title: Home
---

# MailCleaner

> **Warning**: This is an experimental project, not an actual tool. All code in this repository is AI-generated and has not been reviewed for security or correctness. Use at your own risk.

A simple email organizer that matches emails by sender and moves them to folders.

## Features

- **IMAP Support** - Connect to any IMAP server with TLS
- **Flexible Matching** - Match emails by sender address using partial matching
- **Folder Organization** - Automatically move matched emails to specified folders
- **Safe Preview** - Dry-run mode to preview changes before moving emails

## Quick Start

```bash
# Build from source
go build -o mailcleaner

# Preview what would be moved
./mailcleaner -dry-run

# Run and move matching emails
./mailcleaner
```

## Example Configuration

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
    },
    {
      "sender": "@github.com",
      "move_to_folder": "GitHub"
    }
  ]
}
```

## Documentation

- [Getting Started](getting-started) - Installation and initial setup
- [Configuration Reference](configuration) - Full configuration options
- [Usage Examples](usage) - Common use cases and examples

## Source Code

Visit the [GitHub Repository](https://github.com/canburak/mailcleaner) for source code, issues, and contributions.
