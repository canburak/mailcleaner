# MailCleaner

[![Test](https://github.com/mailcleaner/mailcleaner/actions/workflows/test.yml/badge.svg)](https://github.com/mailcleaner/mailcleaner/actions/workflows/test.yml)
[![Security Scan](https://github.com/mailcleaner/mailcleaner/actions/workflows/security.yml/badge.svg)](https://github.com/mailcleaner/mailcleaner/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/mailcleaner/mailcleaner/branch/main/graph/badge.svg)](https://codecov.io/gh/mailcleaner/mailcleaner)
[![Go Report Card](https://goreportcard.com/badge/github.com/mailcleaner/mailcleaner)](https://goreportcard.com/report/github.com/mailcleaner/mailcleaner)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Warning**: This is an experimental project, not an actual tool. All code in this repository is AI-generated and has not been reviewed for security or correctness. Use at your own risk. This code may contain bugs, vulnerabilities, or other issues that could compromise your system or data.

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

### Running the Web Server

```bash
# Build and run the server
go build -o mailcleaner-server ./cmd/server
./mailcleaner-server

# With custom options
./mailcleaner-server -port 8080 -db /path/to/data.db
```

The server runs on `http://localhost:8080` by default.

### Building the Frontend

```bash
cd web
npm install
npm run build

# For development with hot reload
npm run dev
```

### Running with Frontend

```bash
# Build frontend
cd web && npm run build && cd ..

# Serve frontend with API
./mailcleaner-server -static web/dist
```

## API Endpoints

### Accounts
- `GET /api/accounts` - List all accounts
- `POST /api/accounts` - Create account
- `POST /api/accounts/test` - Test connection (without saving)
- `GET /api/accounts/:id` - Get account
- `PUT /api/accounts/:id` - Update account
- `DELETE /api/accounts/:id` - Delete account
- `POST /api/accounts/:id/test` - Test saved account connection
- `GET /api/accounts/:id/folders` - List IMAP folders
- `POST /api/accounts/:id/folders` - Create IMAP folder

### Rules
- `GET /api/accounts/:id/rules` - List rules for account
- `POST /api/accounts/:id/rules` - Create rule
- `GET /api/rules/:id` - Get rule
- `PUT /api/rules/:id` - Update rule
- `DELETE /api/rules/:id` - Delete rule

### Preview
- `GET /api/accounts/:id/preview` - Preview rule matches
- `POST /api/accounts/:id/apply` - Apply rules to move emails
- `WS /ws/preview` - WebSocket for live preview

## CLI Usage

The CLI tool maintains backwards compatibility with the original config format:

```bash
# Build CLI
go build -o mailcleaner ./cmd/mailcleaner

# Preview what would be moved (dry run)
./mailcleaner -dry-run

# Run and move matching emails
./mailcleaner

# Use custom config file
./mailcleaner -config /path/to/config.json
```

### CLI Configuration

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
    },
    {
      "sender": "@github.com",
      "move_to_folder": "GitHub"
    }
  ]
}
```

## Rule Configuration

Rules support three pattern types:

| Type | Description | Example |
|------|-------------|---------|
| `sender` | Match the From address | `newsletter@` matches `newsletter@company.com` |
| `subject` | Match the subject line | `[URGENT]` matches subjects containing that text |
| `from_domain` | Match sender's domain | `github.com` matches all GitHub emails |

All patterns are case-insensitive partial matches.

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

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- SQLite (via CGO)
- Make (optional, but recommended)

### Quick Setup

```bash
# Clone and setup
git clone https://github.com/your-username/mailcleaner.git
cd mailcleaner
./scripts/setup.sh
```

### Using Make

```bash
make help           # Show all available commands
make build          # Build all binaries
make test           # Run tests
make test-cover     # Run tests with coverage
make lint           # Run linters
make dev            # Start development servers
make security-scan  # Run security scanners
```

### Running Tests

```bash
# Using make
make test           # Run all Go tests
make test-cover     # Run with coverage
make test-race      # Run with race detection

# Or directly
go test ./...
go test -cover ./...
go test -race ./...

# Frontend tests
cd web && npm run test
cd web && npm run test:e2e
```

### Docker

```bash
# Build and run with Docker
docker build -t mailcleaner .
docker run -p 8080:8080 mailcleaner

# Or use docker-compose
docker-compose up -d
```

### Database

Data is stored in SQLite at `~/.mailcleaner/data.db` by default. The database schema is automatically migrated on startup.

## Supported Protocols

- **IMAP** - Internet Message Access Protocol (TLS enabled by default, set `"tls": false` for plaintext)

## Deployment

### Deploy to Render (Free)

1. Fork this repository to your GitHub account
2. Go to [Render Dashboard](https://dashboard.render.com)
3. Click **New** > **Blueprint**
4. Connect your GitHub repository
5. Render will auto-detect `render.yaml` and deploy

Or deploy manually:
1. Click **New** > **Web Service**
2. Connect your repository
3. Configure:
   - **Runtime**: Go
   - **Build Command**: `cd web && npm install && npm run build && cd .. && go build -o mailcleaner-server ./cmd/server`
   - **Start Command**: `./mailcleaner-server -static web/dist -db /tmp/data.db`

> **Note**: Render's free tier has no persistent disk. SQLite data is stored in `/tmp` and will reset on deploys/restarts. For persistent data, upgrade to a paid plan with a disk or use an external database.

### Environment Variables

The server supports these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port (set automatically by Render) | `8080` |

## Security Notes

- Passwords are stored in the SQLite database
- Use TLS connections (default) for IMAP servers
- The web server should be run behind a reverse proxy in production
- Consider using OAuth2/XOAUTH2 for Gmail and other providers

## Documentation

For detailed documentation, see the [Documentation Site](docs/index.md):

- [Getting Started](docs/getting-started.md) - Installation and setup
- [Configuration Reference](docs/configuration.md) - Accounts, rules, and settings
- [Usage Examples](docs/usage.md) - Common patterns and automation
- [API Reference](docs/api.md) - REST and WebSocket API documentation

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

- Read the [Code of Conduct](CODE_OF_CONDUCT.md)
- Check existing [issues](https://github.com/mailcleaner/mailcleaner/issues)
- Submit pull requests following our guidelines

## License

MIT
