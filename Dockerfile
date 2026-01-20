# =============================================================================
# Stage 1: Base - System dependencies and user setup
# =============================================================================
FROM ubuntu:24.04 AS base

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

# Install base dependencies
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt/lists \
    apt-get update && apt-get install -y --no-install-recommends \
    zsh \
    bash \
    curl \
    wget \
    git \
    git-lfs \
    vim \
    nano \
    jq \
    ca-certificates \
    locales \
    sudo \
    unzip \
    zip \
    tar \
    rsync \
    less \
    htop \
    lsof \
    net-tools \
    iputils-ping \
    dnsutils \
    tzdata \
    gnupg \
    lsb-release \
    # Runtime libs for language runtimes
    libssl3 \
    zlib1g \
    libyaml-0-2 \
    libreadline8 \
    libncurses6 \
    libffi8 \
    libgdbm6 \
    # Tools requested
    ripgrep \
    fd-find \
    bat \
    # Database clients (lightweight versions)
    sqlite3 \
    && install -d /usr/share/postgresql-common/pgdg \
    && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
        -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
    && . /etc/os-release \
    && echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt ${VERSION_CODENAME}-pgdg main" \
        > /etc/apt/sources.list.d/pgdg.list \
    && curl -fsSL https://packages.redis.io/gpg \
        | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg \
    && chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" \
        > /etc/apt/sources.list.d/redis.list \
    && printf '%s\n' \
        'Package: postgresql-client-18' \
        'Pin: release o=apt.postgresql.org' \
        'Pin-Priority: 1001' \
        > /etc/apt/preferences.d/pgdg \
    && printf '%s\n' \
        'Package: redis-server redis-tools' \
        'Pin: version 6:8.*' \
        'Pin-Priority: 1001' \
        > /etc/apt/preferences.d/redis \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
    # Database clients only (no servers - use docker compose)
    postgresql-client-18 \
    redis-tools \
    && ln -s /usr/bin/fdfind /usr/local/bin/fd \
    && if [ -x /usr/bin/batcat ]; then ln -sf /usr/bin/batcat /usr/local/bin/bat; fi \
    && rm -rf /var/lib/apt/lists/*

# Generate Locale
RUN locale-gen en_US.UTF-8

# Create Agent User
RUN userdel -r ubuntu && \
    useradd -m -s /bin/zsh -u 1000 agent && \
    echo "agent ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Set Claude + Codex config dirs
ENV CLAUDE_CONFIG_DIR=/home/agent/.agents/claude \
    CODEX_HOME=/home/agent/.agents/codex
RUN mkdir -p /home/agent/.agents/claude /home/agent/.agents/codex && \
    chown -R agent:agent /home/agent/.agents

# Install Powerlevel10k & Zsh plugins
USER agent
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.1/zsh-in-docker.sh)" -- \
    -t powerlevel10k \
    -p git \
    -p history \
    -p https://github.com/zsh-users/zsh-autosuggestions \
    -p https://github.com/zsh-users/zsh-syntax-highlighting

# Copy default p10k configuration to avoid the configuration wizard on first run
RUN if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k" ]; then \
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k; \
    fi && \
    cp ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k/config/p10k-rainbow.zsh ~/.p10k.zsh && \
    echo 'typeset -g POWERLEVEL9K_DISABLE_GITSTATUS=true' >> ~/.p10k.zsh && \
    echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> ~/.zshrc

# Pre-download gitstatusd to avoid fetching it on startup
RUN ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k/gitstatus/install

# Fix: Set ZSH_THEME to point to the correct location inside the directory
RUN sed -i 's/ZSH_THEME="powerlevel10k"/ZSH_THEME="powerlevel10k\/powerlevel10k"/' ~/.zshrc

USER root

# =============================================================================
# Stage 1b: Build dependencies (for language compilation)
# =============================================================================
FROM base AS build-deps

RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt/lists \
    apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libssl-dev \
    zlib1g-dev \
    libyaml-dev \
    libreadline-dev \
    libncurses-dev \
    libffi-dev \
    libgdbm-dev \
    && rm -rf /var/lib/apt/lists/*

# =============================================================================
# Stage 2: Languages - Mise and Runtime setup
# =============================================================================
FROM build-deps AS languages

USER agent
ENV HOME=/home/agent
ENV PATH="/home/agent/.local/bin:/home/agent/.local/share/mise/shims:$PATH"

# Install mise
RUN curl https://mise.run | sh

# Copy mise config and install runtimes from it
COPY --chown=agent:agent .config/mise/config.toml /home/agent/.config/mise/config.toml
RUN --mount=type=cache,target=/home/agent/.cache,uid=1000,gid=1000 \
    MISE_TRUSTED_CONFIG=1 mise install

# =============================================================================
# Stage 3: Tools - AI Agents and CLI Tools
# =============================================================================
FROM languages AS tools

ARG CLAUDE_CODE_VERSION="latest"
ARG CODEX_VERSION="latest"

# Install Homebrew
RUN /bin/bash -c "CI=1 /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"" && \
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

# Install Claude Code (native installer)
RUN curl -fsSL https://claude.ai/install.sh | bash

# Install Global NPM Packages (AI Agents)
RUN --mount=type=cache,target=/home/agent/.npm,uid=1000,gid=1000 \
    mise exec node -- npm install -g \
    @openai/codex@latest \
    opencode-ai \
    yarn \
    pnpm \
    eslint \
    prettier

# Create a Gemini wrapper since there isn't a standard 'gemini' CLI yet
RUN mkdir -p /home/agent/.local/bin && \
    echo '#!/bin/bash\nopencode --model "${RIZE_GEMINI_MODEL:-gemini-pro}" "$@"' > /home/agent/.local/bin/gemini && \
    chmod +x /home/agent/.local/bin/gemini

# Install Python tools via pipx
RUN --mount=type=cache,target=/home/agent/.cache/pip,uid=1000,gid=1000 \
    --mount=type=cache,target=/home/agent/.cache/pipx,uid=1000,gid=1000 \
    mise exec python -- pip install pipx && \
    mise exec python -- python -m pipx ensurepath && \
    mise exec python -- pipx install uv && \
    mise exec python -- pipx install yq

# Install Go Tools
# Tools: delve (debug), golangci-lint
RUN --mount=type=cache,target=/home/agent/.cache/go-build,uid=1000,gid=1000 \
    --mount=type=cache,target=/home/agent/go/pkg,uid=1000,gid=1000 \
    mise exec go -- go install github.com/go-delve/delve/cmd/dlv@latest && \
    mise exec go -- go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# =============================================================================
# Stage 5: Slim - Release Image without extra tools
# =============================================================================
FROM base AS slim

COPY --from=languages --chown=agent:agent /home/agent/.local /home/agent/.local
COPY --from=languages --chown=agent:agent /home/agent/.config /home/agent/.config

USER agent
ENV PATH="/home/agent/.local/bin:/home/agent/.local/share/mise/shims:$PATH"
COPY --chown=agent:agent .config/.zshrc /etc/rize/zshrc
RUN echo 'source /etc/rize/zshrc' >> ~/.zshrc

USER root
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/bin/zsh"]
USER root

# =============================================================================
# Stage 4: Final - Release Image
# =============================================================================
FROM base AS final

# Copy Mise and installed runtimes/tools from the 'tools' stage
# Mise stores data in ~/.local/share/mise and ~/.config/mise (or ~/.local/bin/mise for the binary)
COPY --from=tools --chown=agent:agent /home/agent/.local /home/agent/.local
COPY --from=tools --chown=agent:agent /home/agent/.config /home/agent/.config

# Copy Homebrew
COPY --from=tools --chown=agent:agent /home/linuxbrew /home/linuxbrew

# Setup Env for Mise
USER agent
ENV PATH="/home/agent/.local/bin:/home/agent/.local/share/mise/shims:$PATH"
# Ensure mise is activated in zshrc for interactive sessions
# Also configure HISTFILE to use the persisted volume
# And configure Homebrew and Aliases
COPY --chown=agent:agent .config/.zshrc /etc/rize/zshrc
RUN echo 'source /etc/rize/zshrc' >> ~/.zshrc

USER root

# Copy Entrypoint
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Docker Socket permissions workaround (optional, handled by group in entrypoint usually)
# We don't need to do much here as we mount the socket at runtime.

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/bin/zsh"]
USER root
