FROM ubuntu:24.04

# Build-time arguments for multiple versions (comma-separated)
ARG GO_VERSIONS=1.23.3,1.22.5
ARG RUBY_VERSIONS=3.3,3.2
ARG NODE_VERSIONS=24,22
ARG PYTHON_VERSIONS=3.12,3.11
ARG CLAUDE_CODE_VERSION=latest

# Non-interactive setup
ENV DEBIAN_FRONTEND=noninteractive \
    PATH="/home/agent/.local/bin:/home/agent/.local/share/mise/shims:$PATH" \
    HOME=/home/agent \
    USER=agent

# Set the default editor and visual
ENV EDITOR=vim

# Install system dependencies
RUN apt-get update --allow-insecure-repositories --allow-unauthenticated && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    curl \
    wget \
    ca-certificates \
    bash \
    zsh \
    ripgrep \
    fd-find \
    bat \
    jq \
    unzip \
    fzf \
    && ln -s /usr/bin/fdfind /usr/local/bin/fd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user with home directory and make it world-writable
RUN useradd -m -s /bin/bash agent && \
    chmod 777 /home/agent && \
    mkdir -p /home/agent/.local && \
    chmod 777 /home/agent/.local

# Install zsh with powerline10k theme (from Claude Code's devcontainer)
ARG ZSH_IN_DOCKER_VERSION=1.2.0
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v${ZSH_IN_DOCKER_VERSION}/zsh-in-docker.sh)" -- \
  -p git \
  -x

# Make zsh the default shell for agent user
RUN chsh -s /bin/zsh agent

# Add fzf configuration to zshrc if fzf is available
RUN if [ -f /usr/share/doc/fzf/examples/key-bindings.zsh ]; then \
      echo "# FZF configuration" >> /home/agent/.zshrc && \
      echo "source /usr/share/doc/fzf/examples/key-bindings.zsh" >> /home/agent/.zshrc && \
      echo "source /usr/share/doc/fzf/examples/completion.zsh" >> /home/agent/.zshrc; \
    fi

# Install mise using official installer
RUN curl https://mise.run | sh && \
    mv ~/.local/bin/mise /usr/local/bin/mise

# Create workspace directory (as root, before switching user)
RUN mkdir -p /workspace && chmod 777 /workspace

# Create entrypoint script that initializes mise and handles commands
RUN printf '#!/bin/bash\nset -e\ncd /home/agent\n\n# Ensure tools are installed (runs once if needed)\nmise install 2>/dev/null || true\n\n# Activate mise to set up PATH with all tools\neval "$(mise activate bash 2>/dev/null || true)"\n\n# Add npm bin directory to PATH for global packages\nNPM_BIN=$(mise exec node -- npm config get prefix 2>/dev/null)/bin\nif [ -d "$NPM_BIN" ]; then\n  export PATH="$NPM_BIN:$PATH"\nfi\n\nif [ $# -eq 0 ]; then\n    # Interactive shell - use zsh with powerline10k\n    exec /bin/zsh -l\nelse\n    # Execute command with tools available\n    exec "$@"\nfi\n' > /usr/local/bin/docker-entrypoint.sh && \
    chmod +x /usr/local/bin/docker-entrypoint.sh

# Switch to agent user for mise setup
USER agent

# Initialize mise
RUN mise --version

# Create .mise.toml with just the default versions (first one from each list)
# This avoids validation warnings about missing secondary versions
RUN /bin/bash -c 'GO_VERSIONS="${GO_VERSIONS}" && NODE_VERSIONS="${NODE_VERSIONS}" && GO_DEFAULT=$(echo "$GO_VERSIONS" | cut -d, -f1) && NODE_DEFAULT=$(echo "$NODE_VERSIONS" | cut -d, -f1) && printf "[tools]\ngo = \"%s\"\nnode = \"%s\"\n" "$GO_DEFAULT" "$NODE_DEFAULT" > /home/agent/.mise.toml'

# Trust .mise.toml and pre-install default tools
RUN cd /home/agent && \
    mise trust /home/agent/.mise.toml && \
    GO_VERSION=$(echo "$GO_VERSIONS" | cut -d, -f1) && \
    NODE_VERSION=$(echo "$NODE_VERSIONS" | cut -d, -f1) && \
    echo "Pre-installing Go $GO_VERSION and Node $NODE_VERSION..." && \
    mise install go@$GO_VERSION && \
    mise install node@$NODE_VERSION

# Install Claude Code using mise exec
RUN cd /home/agent && \
    mise exec node@24 -- npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}

# Set working directory
WORKDIR /workspace

# Use exec form entrypoint with bash
ENTRYPOINT ["/bin/bash", "/usr/local/bin/docker-entrypoint.sh"]
CMD []
