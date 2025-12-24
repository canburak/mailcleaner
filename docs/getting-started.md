---
layout: default
title: Getting Started
---

# Getting Started

This guide will help you install and configure MailCleaner for your email organization needs.

> **Warning**: This is an experimental project. All code is AI-generated and has not been reviewed for security or correctness. Use at your own risk.

## Prerequisites

- **Go 1.21 or later** - Required to build from source
- **IMAP-enabled email account** - Most email providers support IMAP

## Installation

### Build from Source

Clone the repository and build:

```bash
git clone https://github.com/canburak/mailcleaner.git
cd mailcleaner
go build -o mailcleaner
```

This creates a `mailcleaner` binary in your current directory.

## Configuration

### Create Configuration File

Create a `config.json` file in the same directory as the binary:

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

### Common IMAP Server Settings

| Provider | Server | Port |
|----------|--------|------|
| Gmail | imap.gmail.com | 993 |
| Outlook | outlook.office365.com | 993 |
| Yahoo | imap.mail.yahoo.com | 993 |
| iCloud | imap.mail.me.com | 993 |

**Note**: Some providers require app-specific passwords or enabling "Less Secure Apps" access.

## First Run

### Test with Dry Run

Always test your configuration with a dry run first:

```bash
./mailcleaner -dry-run
```

This shows what emails would be moved without actually moving them.

### Run for Real

Once you're satisfied with the dry run output:

```bash
./mailcleaner
```

## Troubleshooting

### Connection Issues

- Verify your server and port settings
- Check that your password is correct
- Ensure IMAP is enabled in your email provider settings
- For Gmail, you may need an [App Password](https://support.google.com/accounts/answer/185833)

### Folder Not Found

- The destination folder must already exist in your mailbox
- Folder names are case-sensitive on some servers
- Use the full path for nested folders (e.g., `INBOX/Newsletters`)

## Next Steps

- [Configuration Reference](configuration) - Learn about all configuration options
- [Usage Examples](usage) - See common use cases and patterns
