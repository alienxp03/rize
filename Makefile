.PHONY: help build build-cli build-image install test clean push

IMAGE_NAME := alienxp03/rize:latest
BINARY_NAME := rize
BUILD_DIR := ./bin

# Default target
help:
	@echo "Rize - Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  make build              Build both CLI and Docker image"
	@echo "  make build-cli          Build the Go CLI binary"
	@echo "  make build-image        Build the Docker image"
	@echo "  make test               Run tests"
	@echo "  make clean              Remove build artifacts"
	@echo ""
	@echo "Installation:"
	@echo "  make install            Install rize to /usr/local/bin"
	@echo ""
	@echo "Docker:"
	@echo "  make push               Push the Docker image to Docker Hub"
	@echo ""

# Build both CLI and Docker image
build: build-cli build-image

# Build the Go CLI binary
build-cli:
	@echo "Building Go CLI binary..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/rize
	@echo "✓ Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/rize
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/rize
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/rize
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/rize
	@echo "✓ Multi-platform binaries built"

# Build the Docker image
build-image:
	@echo "Building Docker image $(IMAGE_NAME)..."
	@docker build -t $(IMAGE_NAME) .
	@echo "✓ Image built"

# Build and push Docker image
build-push: build-image
	@echo "Pushing to Docker Hub..."
	@docker push $(IMAGE_NAME)
	@echo "✓ Push complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"

# Install rize to /usr/local/bin
install: build-cli
	@echo "Installing rize to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "✓ Installed to /usr/local/bin/$(BINARY_NAME)"

# Push the Docker image to Docker Hub
push:
	@echo "Pushing $(IMAGE_NAME) to Docker Hub..."
	@docker push $(IMAGE_NAME)
	@echo "✓ Push complete"
