# =============================================================================
# Stage 1: Base - System dependencies, user setup, mise, zsh
# This stage rarely changes and provides the foundation
# =============================================================================
FROM ubuntu:24.04 AS base

# Build-time arguments (set via --build-arg, no defaults here)
ARG GO_VERSIONS
ARG NODE_VERSIONS
ARG RUBY_VERSIONS
ARG PYTHON_VERSIONS
ARG NODE_DEFAULT
ARG CLAUDE_CODE_VERSION
ARG CODEX_VERSION

# Non-interactive setup
ENV DEBIAN_FRONTEND=noninteractive \
    PATH="/home/agent/.linuxbrew/bin:/home/agent/.linuxbrew/sbin:/home/agent/.local/bin:/home/agent/.local/go/bin:/home/agent/.local/share/mise/shims:$PATH" \
    HOME=/home/agent \
    USER=agent \
    IS_SANDBOX=1 \
    SHELL=/bin/zsh \
    EDITOR=vim \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    ZSH_DISABLE_COMPFIX=true

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    vim \
    nano \
    less \
    build-essential \
    git \
    git-lfs \
    curl \
    wget \
    httpie \
    locales \
    ca-certificates \
    bash \
    zsh \
    ripgrep \
    fd-find \
    bat \
    jq \
    unzip \
    fzf \
    yq \
    tree \
    dnsutils \
    gh \
    glab \
    sqlite3 \
    postgresql-client \
    redis-tools \
    mysql-client \
    zip \
    tar \
    rsync \
    htop \
    lsof \
    file \
    python3 \
    python3-pip \
    python3-venv \
    python3-virtualenv \
    pipx \
    cargo \
    libssl-dev \
    zlib1g-dev \
    libreadline-dev \
    libyaml-dev \
    libpq-dev \
    sudo \
    && ln -s /usr/bin/fdfind /usr/local/bin/fd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Generate UTF-8 locale
RUN locale-gen en_US.UTF-8 && update-locale LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8

# Create non-root user with home directory
RUN if id ubuntu >/dev/null 2>&1; then userdel -r ubuntu; fi && \
    useradd -m -u 1000 -s /bin/bash agent && \
    mkdir -p /home/agent/workspace /home/agent/.claude-config /home/agent/.local && \
    chown -R agent:agent /home/agent && \
    chmod 755 /home/agent && \
    echo "agent ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install zsh with powerline10k theme (from Claude Code's devcontainer)
ARG ZSH_IN_DOCKER_VERSION=1.2.1
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v${ZSH_IN_DOCKER_VERSION}/zsh-in-docker.sh)" -- \
  -x -u agent

# Fix oh-my-zsh directory permissions (remove group/other write)
RUN find /home/agent/.oh-my-zsh -type d -exec chmod go-w {} \; 2>/dev/null || true

# Disable git status in powerlevel10k
RUN printf 'typeset -g POWERLEVEL9K_DISABLE_GITSTATUS=true\n' >> /home/agent/.zshrc

# Add fzf configuration to zshrc if fzf is available
RUN if [ -f /usr/share/doc/fzf/examples/key-bindings.zsh ]; then \
      echo "# FZF configuration" >> /home/agent/.zshrc && \
      echo "source /usr/share/doc/fzf/examples/key-bindings.zsh" >> /home/agent/.zshrc && \
      echo "source /usr/share/doc/fzf/examples/completion.zsh" >> /home/agent/.zshrc; \
    fi

# Install mise using official installer
RUN curl https://mise.run | sh && \
    mv /home/agent/.local/bin/mise /usr/local/bin/mise

# Final home ownership fix before switching users
RUN chown -R agent:agent /home/agent

# Install Homebrew in standard location for binary support
RUN mkdir -p /home/linuxbrew/.linuxbrew && \
    chown -R agent:agent /home/linuxbrew && \
    chmod -R 755 /home/linuxbrew

# Switch to agent user for subsequent installations
USER agent

# Install Homebrew as agent (non-interactive)
RUN git clone https://github.com/Homebrew/brew /home/linuxbrew/.linuxbrew/Homebrew && \
    mkdir -p /home/linuxbrew/.linuxbrew/bin && \
    ln -s ../Homebrew/bin/brew /home/linuxbrew/.linuxbrew/bin/brew && \
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)" && \
    brew update --force --quiet && \
    cat <<'BREWRC' >> ~/.zshrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
BREWRC
RUN cat <<'BREWRC' >> ~/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
BREWRC

# =============================================================================
# Stage 2: Languages - Pre-install all language versions via mise
# This stage caches language installations separately from tools
# =============================================================================
FROM base AS languages

# Initialize mise
RUN mise --version

# Set global defaults for Ruby
RUN /bin/bash -c 'cd /home/agent; RUBY_DEFAULT=$(echo "$RUBY_VERSIONS" | cut -d, -f1); echo "Setting Ruby default: ${RUBY_DEFAULT}"; mise use -g ruby@"$RUBY_DEFAULT"'

# Pre-install all requested Ruby versions (done separately to avoid timeout)
RUN /bin/bash -c 'IFS="," read -ra RUBY_LIST <<< "$RUBY_VERSIONS"; echo "Pre-installing Ruby versions: ${RUBY_LIST[*]}"; for v in "${RUBY_LIST[@]}"; do mise install ruby@"$v"; done'

# Set global default for Python and pre-install versions
RUN /bin/bash -c 'cd /home/agent; PYTHON_DEFAULT=$(echo "$PYTHON_VERSIONS" | cut -d, -f1); echo "Setting Python default: ${PYTHON_DEFAULT}"; mise use -g python@"$PYTHON_DEFAULT"'

# Pre-install all requested Python versions (done separately to avoid timeout)
RUN /bin/bash -c 'IFS="," read -ra PYTHON_LIST <<< "$PYTHON_VERSIONS"; echo "Pre-installing Python versions: ${PYTHON_LIST[*]}"; for v in "${PYTHON_LIST[@]}"; do mise install python@"$v"; done'

# Set global default for Go and pre-install versions
RUN /bin/bash -c 'cd /home/agent; GO_DEFAULT=$(echo "$GO_VERSIONS" | cut -d, -f1); echo "Setting Go default: ${GO_DEFAULT}"; mise use -g go@"$GO_DEFAULT"; IFS="," read -ra GO_LIST <<< "$GO_VERSIONS"; echo "Pre-installing Go versions: ${GO_LIST[*]}"; for v in "${GO_LIST[@]}"; do mise install go@"$v"; done'

# Set global default for Node and pre-install versions
RUN /bin/bash -c 'cd /home/agent; NODE_DEFAULT="${NODE_DEFAULT}"; echo "Setting Node default: ${NODE_DEFAULT}"; mise use -g node@"$NODE_DEFAULT"; echo "Pre-installing default Node version: ${NODE_DEFAULT}"; mise install node@"$NODE_DEFAULT"; IFS="," read -ra NODE_LIST <<< "$NODE_VERSIONS"; echo "Pre-installing Node versions: ${NODE_LIST[*]}"; for v in "${NODE_LIST[@]}"; do mise install node@"$v"; done'

# =============================================================================
# Stage 3: Tools - Install development tools (npm packages, Go tools, etc.)
# This stage caches tool installations separately from entrypoint changes
# =============================================================================
FROM languages AS tools

# Install Claude Code and Codex using the default Node version
RUN NODE_DEFAULT="${NODE_DEFAULT}" && \
    cd /home/agent && \
    mise exec node@$NODE_DEFAULT -- npm install -g \
      @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION} \
      @openai/codex@${CODEX_VERSION} \
      opencode-ai \
      pnpm \
      yarn \
      eslint \
      prettier

# Install Go tools (user bin to avoid permission issues)
RUN GO_DEFAULT=$(echo "$GO_VERSIONS" | cut -d, -f1) && \
    mkdir -p /home/agent/.local/go/bin && \
    GOBIN=/home/agent/.local/go/bin mise exec go@$GO_DEFAULT -- go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    GOBIN=/home/agent/.local/go/bin mise exec go@$GO_DEFAULT -- go install github.com/go-delve/delve/cmd/dlv@latest

# Install Ruby tools
RUN RUBY_DEFAULT=$(echo "$RUBY_VERSIONS" | cut -d, -f1) && \
    mise exec ruby@$RUBY_DEFAULT -- gem install --no-document bundler rubocop

# Install Python tools via pipx
RUN pipx ensurepath && \
    pipx install uv && \
    pipx install yq || true

# =============================================================================
# Stage 4: Final - Entrypoint script and configuration
# This stage is rebuilt when the rize script changes, but tools remain cached
# =============================================================================
FROM tools AS final

# Switch back to root for entrypoint creation
USER root

# Create entrypoint script that initializes mise and handles commands
# NOTE: This is the ONLY layer that rebuilds when the rize script is modified
RUN printf '#!/bin/bash\nset -e\nexport MISE_TRUSTED_CONFIG=1\n# Use dynamic workspace path from environment, or default to old behavior\nRIZE_WORKSPACE_PATH="${RIZE_WORKSPACE_PATH:-/home/agent/workspace}"\n# Ensure workspace directory exists and has correct permissions\nif [ ! -d "$RIZE_WORKSPACE_PATH" ]; then\n  mkdir -p "$RIZE_WORKSPACE_PATH"\nfi\nif [ "$(stat -c %%u "$RIZE_WORKSPACE_PATH")" != "$(id -u agent)" ]; then\n  sudo chown agent:agent "$RIZE_WORKSPACE_PATH"\nfi\ncd "$RIZE_WORKSPACE_PATH"\n' > /usr/local/bin/docker-entrypoint.sh && \
    printf 'if [ -f /home/agent/.env ]; then\n  set -a\n  source /home/agent/.env\n  set +a\nfi\nif [ -f ./.mise.toml ]; then\n  mise trust ./.mise.toml 2>/dev/null || true\nfi\nmise install 2>/dev/null || true\neval "$(mise activate bash 2>/dev/null || true)"\nNPM_PREFIX=$(mise exec node -- npm config get prefix 2>/dev/null || echo "")\nif [ -n "$NPM_PREFIX" ]; then\n  NPM_BIN="$NPM_PREFIX/bin"\n  if [ -d "$NPM_BIN" ]; then\n  export PATH="$NPM_BIN:$PATH"\n  fi\nfi\nif [ $# -eq 0 ]; then\n    exec /bin/zsh -l\nelse\n    exec "$@"\nfi\n' >> /usr/local/bin/docker-entrypoint.sh && chmod +x /usr/local/bin/docker-entrypoint.sh

# Set working directory
WORKDIR /home/agent

# Use exec form entrypoint with bash
ENTRYPOINT ["/bin/bash", "/usr/local/bin/docker-entrypoint.sh"]
CMD []
