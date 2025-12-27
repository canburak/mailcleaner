# Contributing to MailCleaner

Thank you for your interest in contributing! Please read our guidelines below.

## Code of Conduct

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- SQLite (via CGO)

### Quick Start

```bash
# Clone and setup
git clone https://github.com/your-username/mailcleaner.git
cd mailcleaner
./scripts/setup.sh

# Or manually
go mod download
cd web && npm ci && cd ..
pre-commit install
```

### Common Commands

```bash
make build          # Build all binaries
make test           # Run tests
make lint           # Run linters
make dev            # Start dev servers
```

## Commit Messages

We use [Conventional Commits](https://conventionalcommits.org/):

```
<type>(<scope>): <description>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `ci`, `chore`, `deps`, `sec`

Examples:
```
feat(api): add bulk delete endpoint
fix(imap): handle timeout gracefully
docs: update API reference
```

## Pull Request Process

1. Fork and create a branch
2. Make changes and add tests
3. Run `make test lint`
4. Submit PR with clear description

## Testing

```bash
make test           # Go tests
make test-cover     # With coverage
make test-web       # Frontend tests
make test-e2e       # E2E tests
```

## Security

Report vulnerabilities via [Security Policy](SECURITY.md) - not public issues.

## License

Contributions are licensed under the MIT License.
