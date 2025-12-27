.PHONY: all build build-server build-cli build-web test test-cover test-race lint lint-go lint-web fmt vet clean install-tools dev run-server docker-build docker-run help security-scan

# Variables
BINARY_NAME=mailcleaner
SERVER_BINARY=mailcleaner-server
GO=go
GOFLAGS=-v
LDFLAGS=-s -w
COVERAGE_FILE=coverage.out
DOCKER_IMAGE=mailcleaner

# Default target
all: lint test build

# Build targets
build: build-cli build-server ## Build all binaries

build-cli: ## Build CLI binary
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/mailcleaner

build-server: ## Build server binary
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(SERVER_BINARY) ./cmd/server

build-web: ## Build frontend
	cd web && npm ci && npm run build

# Test targets
test: ## Run tests
	$(GO) test $(GOFLAGS) ./...

test-cover: ## Run tests with coverage
	$(GO) test $(GOFLAGS) -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

test-cover-html: test-cover ## Generate HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html

test-race: ## Run tests with race detector
	$(GO) test $(GOFLAGS) -race ./...

test-all: test-race test-cover ## Run all tests with race detection and coverage

test-web: ## Run frontend tests
	cd web && npm run test

test-web-coverage: ## Run frontend tests with coverage
	cd web && npm run test:coverage

test-e2e: ## Run E2E tests
	cd web && npm run test:e2e

# Lint targets
lint: lint-go lint-web ## Run all linters

lint-go: ## Run Go linters
	golangci-lint run ./...

lint-web: ## Run frontend linters
	cd web && npm run lint

lint-fix: ## Fix linting issues
	golangci-lint run --fix ./...
	cd web && npm run lint:fix

# Format targets
fmt: ## Format Go code
	$(GO) fmt ./...

fmt-web: ## Format frontend code
	cd web && npm run format

fmt-all: fmt fmt-web ## Format all code

# Vet target
vet: ## Run go vet
	$(GO) vet ./...

# Security targets
security-scan: ## Run security scanners
	gosec ./...
	cd web && npm audit

security-scan-sarif: ## Run security scanners with SARIF output
	gosec -fmt sarif -out gosec-results.sarif ./...

# Clean target
clean: ## Clean build artifacts
	rm -f $(BINARY_NAME) $(SERVER_BINARY)
	rm -f $(COVERAGE_FILE) coverage.html
	rm -rf web/dist web/node_modules
	rm -f gosec-results.sarif

# Install development tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/goreleaser/goreleaser@latest
	cd web && npm ci

# Development targets
dev: ## Run development server (frontend + backend)
	@echo "Starting backend server..."
	$(GO) run ./cmd/server -port 8080 &
	@echo "Starting frontend dev server..."
	cd web && npm run dev

run-server: ## Run the server
	$(GO) run ./cmd/server

# Docker targets
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 $(DOCKER_IMAGE)

docker-compose-up: ## Start with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	docker-compose down

# Release targets
release-snapshot: ## Create a snapshot release (no publish)
	goreleaser release --snapshot --clean

release: ## Create a release
	goreleaser release --clean

# Pre-commit
pre-commit-install: ## Install pre-commit hooks
	pre-commit install
	pre-commit install --hook-type commit-msg

pre-commit-run: ## Run pre-commit on all files
	pre-commit run --all-files

# Database
db-reset: ## Reset the database
	rm -f ~/.mailcleaner/data.db

# Help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
