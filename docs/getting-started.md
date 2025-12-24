---
layout: default
title: Getting Started
---

# Getting Started

This guide will help you install and configure MailCleaner for your email organization needs.

> **Warning**: This is an experimental project. All code is AI-generated and has not been reviewed for security or correctness. Use at your own risk.

## Prerequisites

- **Go 1.21 or later** - Required to build the server and CLI
- **Node.js 18+** - Required to build the web frontend
- **SQLite** - Required for data persistence (via CGO)
- **IMAP-enabled email account** - Most email providers support IMAP

## Installation

### Build from Source

Clone the repository:

```bash
git clone https://github.com/canburak/mailcleaner.git
cd mailcleaner
```

### Build the Web Server

```bash
go build -o mailcleaner-server ./cmd/server
```

### Build the Frontend

```bash
cd web
npm install
npm run build
cd ..
```

### Build the CLI Tool

```bash
go build -o mailcleaner ./cmd/mailcleaner
```

## Running the Web Server

### Basic Usage

```bash
./mailcleaner-server
```

The server runs on `http://localhost:8080` by default.

### With Frontend

```bash
./mailcleaner-server -static web/dist
```

### Server Options

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | HTTP server port | `8080` |
| `-db` | Database file path | `~/.mailcleaner/data.db` |
| `-static` | Static files directory | (none) |

### Example

```bash
./mailcleaner-server -port 3000 -db /var/lib/mailcleaner/data.db -static web/dist
```

## First Steps with Web UI

1. **Add an Account** - Navigate to Accounts and click "Add Account"
2. **Configure IMAP Settings** - Enter your email server details
3. **Test Connection** - Verify the connection works
4. **Create Rules** - Define patterns to match emails
5. **Preview** - See which emails match your rules
6. **Apply** - Move matched emails to their destination folders

## Common IMAP Server Settings

| Provider | Server | Port |
|----------|--------|------|
| Gmail | imap.gmail.com | 993 |
| Outlook | outlook.office365.com | 993 |
| Yahoo | imap.mail.yahoo.com | 993 |
| iCloud | imap.mail.me.com | 993 |

**Note**: Some providers require app-specific passwords or OAuth2.

## CLI Quick Start

For automation or scripting, use the CLI tool:

```bash
# Create config file
cat > config.json << 'EOF'
{
  "server": "imap.gmail.com",
  "port": 993,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "rules": [
    {
      "sender": "@github.com",
      "move_to_folder": "GitHub"
    }
  ]
}
EOF

# Preview what would be moved
./mailcleaner -dry-run

# Apply rules
./mailcleaner
```

## Development Mode

For frontend development with hot reload:

```bash
# Terminal 1: Run the API server
./mailcleaner-server

# Terminal 2: Run frontend dev server
cd web
npm run dev
```

The frontend dev server proxies API requests to the backend.

## Next Steps

- [Configuration Reference](configuration) - Learn about all configuration options
- [Usage Examples](usage) - See common use cases and patterns
- [API Reference](api) - REST API documentation
