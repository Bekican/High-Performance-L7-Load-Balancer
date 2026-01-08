# Makefile for Go Load Balancer Project

# Variables
BINARY_NAME=loadbalancer
DOCKER_IMAGE=go-loadbalancer
MAIN_PATH=cmd/lb/main.go

.PHONY: all build test clean run docker-build docker-run

all: build

# Build the binary
build:
	@echo "Building..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application locally
run: build
	@echo "Running..."
	./$(BINARY_NAME)

# Clean up binaries
clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose..."
	docker-compose down

# Helper for PowerShell users (if make is not installed, they can read this)
help:
	@echo "Available commands:"
	@echo "  make build        - Build the Go binary"
	@echo "  make test         - Run unit tests"
	@echo "  make run          - Build and run the application"
	@echo "  make docker-build - Build the Docker image"
	@echo "  make docker-run   - Start services with Docker Compose"
