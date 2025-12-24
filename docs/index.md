---
layout: default
title: Home
---

# MailCleaner

> **Warning**: This is an experimental project, not an actual tool. All code in this repository is AI-generated and has not been reviewed for security or correctness. Use at your own risk.

A full-stack email organizer with a web UI for managing email filtering rules and live IMAP preview.

## Features

- **Web Interface** - Modern Vue 3 frontend for managing accounts and rules
- **Multiple IMAP Accounts** - Manage multiple email accounts from one interface
- **Live Preview** - Test rules against live emails via WebSocket before applying
- **Flexible Rules** - Match by sender, subject, or domain with pattern matching
- **Rule Priority** - Control which rules apply first with priority ordering
- **CLI Support** - Command-line tool for automation and scripting

## Architecture

```
mailcleaner/
├── cmd/
│   ├── mailcleaner/     # CLI tool
│   └── server/          # Web server
├── internal/
│   ├── api/             # REST API handlers
│   ├── imap/            # IMAP client library
│   ├── models/          # Data structures
│   └── storage/         # SQLite persistence
├── web/                 # Vue 3 frontend
└── testserver/          # In-memory IMAP test server
```

## Quick Start

### Web Server

```bash
# Build and run the server
go build -o mailcleaner-server ./cmd/server
./mailcleaner-server

# With frontend
cd web && npm install && npm run build && cd ..
./mailcleaner-server -static web/dist
```

The server runs on `http://localhost:8080` by default.

### CLI Tool

```bash
# Build CLI
go build -o mailcleaner ./cmd/mailcleaner

# Preview what would be moved
./mailcleaner -dry-run

# Run and move matching emails
./mailcleaner
```

## Documentation

- [Getting Started](getting-started) - Installation and initial setup
- [Configuration Reference](configuration) - Rules, accounts, and settings
- [Usage Examples](usage) - Common use cases for web UI and CLI
- [API Reference](api) - REST API and WebSocket documentation

## Source Code

Visit the [GitHub Repository](https://github.com/canburak/mailcleaner) for source code, issues, and contributions.
