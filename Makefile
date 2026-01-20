.PHONY: help build install push exec

IMAGE_NAME := alienxp03/rize:latest

# Default target
help:
	@echo "Rize Dev Container - Available targets:"
	@echo ""
	@echo "  make build              Build the Docker image (for maintainers)"
	@echo "  make push               Push the Docker image to Docker Hub"
	@echo "  make install            Install rize to ~/.local/bin/"
	@echo "  make exec <cmd...>      Run a command via rize exec"
	@echo ""

build-local:
	@echo "Building Docker image $(IMAGE_NAME)..."
	docker build -t $(IMAGE_NAME) .

build-push:
	@echo "Building Docker image $(IMAGE_NAME)..."
	docker build -t $(IMAGE_NAME) .
	@echo "Pushing to Docker Hub..."
	docker push $(IMAGE_NAME)

install:
	@./rize install

exec:
	@./rize exec $(filter-out $@,$(MAKECMDGOALS))

# Ignore extra arguments passed as make goals.
%:
	@:

# Push the Docker image to Docker Hub
push:
	@echo "Pushing alienxp03/rize:latest to Docker Hub..."
	@docker push alienxp03/rize:latest
	@echo "✓ Push complete"
