# Makefile

# Build variables
BINARY_NAME=proxy-service
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"



# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER_BUILD=docker build

# Kubernetes parameters
KUBECTL=kubectl

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# Environment variables
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

.PHONY: all build clean test coverage deps lint docker-build docker-run k8s-deploy

# Default target
all: clean build test

# Build the application
build:
    go build $(LDFLAGS) -o bin/proxy-service ./cmd/proxy-service
# build:
# 	@echo "Building $(BINARY_NAME)..."
# 	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/proxy-service

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.*
	$(DOCKER_COMPOSE) down -v

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Generate test coverage
coverage:
	@echo "Generating test coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Generate certificates
generate-certs:
	@echo "Generating certificates..."
	mkdir -p cert
	openssl req -x509 -newkey rsa:4096 \
		-keyout cert/server.key \
		-out cert/server.crt \
		-days 365 -nodes \
		-subj "/CN=proxy-service"

# Docker targets
docker-build:
	@echo "Building Docker image..."
	$(DOCKER_BUILD) -t $(BINARY_NAME):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-f deployment/docker/Dockerfile .

docker-run:
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) up -d

docker-stop:
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down

docker-logs:
	@echo "Showing Docker logs..."
	$(DOCKER_COMPOSE) logs -f

# Development targets
dev: docker-run
	@echo "Starting development environment..."
	air -c .air.toml

# Kubernetes targets
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	$(KUBECTL) apply -f deployment/kubernetes/

k8s-delete:
	@echo "Removing from Kubernetes..."
	$(KUBECTL) delete -f deployment/kubernetes/

k8s-logs:
	@echo "Showing Kubernetes logs..."
	$(KUBECTL) logs -f deployment/$(BINARY_NAME)

# Monitoring setup
setup-monitoring:
	@echo "Setting up monitoring..."
	./scripts/monitoring-setup.sh

# Database targets
db-migrate:
	@echo "Running database migrations..."
	$(GORUN) cmd/migrate/main.go

db-seed:
	@echo "Seeding database..."
	$(GORUN) cmd/seed/main.go

# Help target
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  coverage       - Generate test coverage report"
	@echo "  deps           - Download dependencies"
	@echo "  lint           - Run linter"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Start Docker containers"
	@echo "  docker-stop    - Stop Docker containers"
	@echo "  docker-logs    - Show Docker logs"
	@echo "  dev            - Start development environment"
	@echo "  k8s-deploy     - Deploy to Kubernetes"
	@echo "  k8s-delete     - Remove from Kubernetes"
	@echo "  k8s-logs       - Show Kubernetes logs"
	@echo "  db-migrate     - Run database migrations"
	@echo "  db-seed        - Seed database"