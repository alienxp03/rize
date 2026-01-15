.PHONY: help build install push

# Default target
help:
	@echo "Rize Dev Container - Available targets:"
	@echo ""
	@echo "  make build              Build the Docker image (for maintainers)"
	@echo "  make push               Push the Docker image to Docker Hub"
	@echo "  make install            Install rize to ~/.local/bin/"
	@echo ""

# Build the Docker image with default versions (for maintainers)
build:
	@echo "Extracting Dockerfile from rize script..."
	@sed -n '/^DOCKERFILE_CONTENT=/,/^DOCKERFILE_EOF$$/p' rize | \
	 sed '1d;$$d' > Dockerfile
	@echo "Building Docker image..."
	docker build -t alienxp03/rize:latest \
		--build-arg "GO_VERSIONS=1.25.5,1.23.3" \
		--build-arg "RUBY_VERSIONS=3.4.7,3.1.7" \
		--build-arg "PYTHON_VERSIONS=3.13.0" \
		--build-arg "NODE_VERSIONS=24" \
		--build-arg "NODE_DEFAULT=24" \
		--build-arg "CLAUDE_CODE_VERSION=latest" \
		--build-arg "CODEX_VERSION=latest" \
		.
	@docker tag alienxp03/rize:latest rize:latest
	@echo "✓ Build complete. Tagged as alienxp03/rize:latest and rize:latest"

# Install rize to ~/.local/bin
install:
	@echo "Installing rize to ~/.local/bin/"
	@mkdir -p ~/.local/bin
	@cp rize ~/.local/bin/rize
	@chmod +x ~/.local/bin/rize
	@echo "✓ Installation complete"
	@echo "Run 'rize shell' to get started"

# Push the Docker image to Docker Hub
push:
	@echo "Pushing alienxp03/rize:latest to Docker Hub..."
	@docker push alienxp03/rize:latest
	@echo "✓ Push complete"

