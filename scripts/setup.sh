#!/bin/bash
set -e

echo "=== MailCleaner Development Setup ==="
echo ""

# Check for required tools
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "Error: $1 is required but not installed."
        exit 1
    fi
}

echo "Checking prerequisites..."
check_command go
check_command node
check_command npm

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
NODE_VERSION=$(node --version | sed 's/v//')
echo "  Go version: $GO_VERSION"
echo "  Node version: $NODE_VERSION"
echo ""

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download
echo "  Done!"
echo ""

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd web
npm ci
cd ..
echo "  Done!"
echo ""

# Install development tools
echo "Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
echo "  Done!"
echo ""

# Install pre-commit hooks
if command -v pre-commit &> /dev/null; then
    echo "Installing pre-commit hooks..."
    pre-commit install
    pre-commit install --hook-type commit-msg
    echo "  Done!"
else
    echo "Warning: pre-commit not found. Install it with: pip install pre-commit"
fi
echo ""

# Build the project
echo "Building the project..."
go build -o mailcleaner ./cmd/mailcleaner
go build -o mailcleaner-server ./cmd/server
echo "  Done!"
echo ""

echo "=== Setup Complete ==="
echo ""
echo "Quick start:"
echo "  make run-server    # Start the backend server"
echo "  cd web && npm run dev  # Start the frontend dev server"
echo ""
echo "Run 'make help' to see all available commands."
