# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Email the maintainers directly or use GitHub's private vulnerability reporting feature
3. Include as much detail as possible:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Assessment**: We will assess the vulnerability and determine its severity
- **Updates**: We will keep you informed of our progress
- **Resolution**: We aim to resolve critical issues within 7 days
- **Credit**: We will credit reporters in our release notes (unless you prefer anonymity)

## Security Measures

This project implements the following security measures:

### Automated Security Scanning

- **CodeQL**: Static analysis for Go and JavaScript/TypeScript
- **gosec**: Go-specific security scanner
- **Trivy**: Vulnerability scanning for dependencies and filesystem
- **Gitleaks**: Secret detection in git history
- **npm audit**: Node.js dependency vulnerability scanning
- **Dependabot**: Automated dependency updates

### Code Quality

- Pre-commit hooks for code quality checks
- golangci-lint for comprehensive Go linting
- ESLint and Prettier for frontend code

### Runtime Security

- TLS enabled by default for IMAP connections
- Password fields excluded from API responses
- CORS middleware configured
- Input validation on API endpoints

## Security Considerations

### Password Storage

> **Warning**: Passwords are currently stored in plaintext in the SQLite database.

For production use, consider:
- Using OAuth2/XOAUTH2 for Gmail and other providers
- Running behind a reverse proxy with TLS
- Restricting access to the database file
- Using environment variables for sensitive configuration

### Network Security

- Always use TLS for IMAP connections (default)
- Run the web server behind a reverse proxy (nginx, Caddy) in production
- Use HTTPS for all web traffic
- Restrict access to trusted networks

### Development

- Never commit secrets or credentials
- Use `.env` files for local development (gitignored)
- Run security scans before merging PRs

## Dependencies

We regularly update dependencies to patch known vulnerabilities:

- Go dependencies: Weekly via Dependabot
- npm dependencies: Weekly via Dependabot
- GitHub Actions: Weekly via Dependabot

## Contact

For security concerns, please use GitHub's private vulnerability reporting feature or contact the maintainers directly.
