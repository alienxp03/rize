#!/bin/bash
set -euo pipefail

# Defaults
HOST_UID=${HOST_UID:-1000}
HOST_GID=${HOST_GID:-1000}
USERNAME=agent
WORKSPACE_DIR=${RIZE_WORKSPACE_DIR:-/workspace}
export WORKSPACE_DIR

# Use sudo when not running as root
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
fi

# Check if we need to adjust the user's UID/GID
if [ "$(id -u $USERNAME)" != "$HOST_UID" ] || [ "$(id -g $USERNAME)" != "$HOST_GID" ]; then
    # echo "Updating UID/GID to $HOST_UID:$HOST_GID..."

    # Update GID if needed
    if [ "$(id -g $USERNAME)" != "$HOST_GID" ]; then
        # Check if group already exists
        if ! getent group "$HOST_GID" >/dev/null; then
            # Only change the agent group GID if the target GID doesn't exist
            $SUDO groupmod -g "$HOST_GID" "$USERNAME"
        fi
    fi

    # Update UID
    $SUDO usermod -u "$HOST_UID" -g "$HOST_GID" "$USERNAME"

    # Fix permissions for key paths only (avoid full home scan)
    $SUDO chown "$HOST_UID:$HOST_GID" "/home/$USERNAME"
    for path in \
        "/home/$USERNAME/.local" \
        "/home/$USERNAME/.config" \
        "/home/$USERNAME/.cache" \
        "/home/$USERNAME/.npm" \
        "/home/$USERNAME/.cargo" \
        "/home/$USERNAME/.rustup" \
        "/home/$USERNAME/.gem" \
        "/home/$USERNAME/.bundle" \
        "/home/$USERNAME/.go"
    do
        if [ -e "$path" ]; then
            $SUDO chown -R "$HOST_UID:$HOST_GID" "$path" 2>/dev/null || true
        fi
    done
fi

# Ensure workspace (cwd) is writable if it's not mounted
if [ ! -d "$WORKSPACE_DIR" ]; then
    $SUDO mkdir -p "$WORKSPACE_DIR"
    $SUDO chown "$USERNAME:$USERNAME" "$WORKSPACE_DIR"
fi

# Switch to the user for the remaining commands
if [ "$1" = "exec_as_root" ]; then
    shift
    if [ "$(id -u)" -ne 0 ]; then
        exec sudo "$@"
    else
        exec "$@"
    fi
else
    # Ensure agent dirs are present and writable
    if [ "$(id -u)" -eq 0 ]; then
        mkdir -p /home/agent/.agents/claude /home/agent/.agents/codex
        chown -R agent:agent /home/agent/.agents
    fi

    # Prepare the environment command
    # Initialize mise in the environment before running the command
    # We wrap the command in a shell that sources the environment

    if [ "$(id -u)" -eq 0 ]; then
        exec sudo -E -H -u "$USERNAME" bash -lc '
        export PATH="/home/agent/.local/bin:$PATH"
        export MISE_TRUSTED_CONFIG=1
        if [ -f /home/agent/.env ]; then
            set -a
            source /home/agent/.env
            set +a
        fi

        # Source .rize/env if it exists
        if [ -f /home/agent/.rize/env ]; then
            set -a
            source /home/agent/.rize/env
            set +a
        fi

        # Activate mise
        if [ -f "$WORKSPACE_DIR/.config/mise/config.toml" ]; then
            mise trust "$WORKSPACE_DIR/.config/mise/config.toml" >/dev/null 2>&1 || true
        fi
        eval "$(mise activate bash)"

        # Add npm binaries to PATH if they exist
        NPM_PREFIX=$(npm config get prefix 2>/dev/null || echo "")
        if [ -n "$NPM_PREFIX" ]; then
            export PATH="$NPM_PREFIX/bin:$PATH"
        fi

        # Run the command
        if [ $# -eq 0 ]; then
            exec /bin/zsh -l
        else
            exec "$@"
        fi
        ' -- "$@"
    else
        exec bash -lc '
        export PATH="/home/agent/.local/bin:$PATH"
        export MISE_TRUSTED_CONFIG=1
        if [ -f /home/agent/.env ]; then
            set -a
            source /home/agent/.env
            set +a
        fi

        # Source .rize/env if it exists
        if [ -f /home/agent/.rize/env ]; then
            set -a
            source /home/agent/.rize/env
            set +a
        fi

        # Activate mise
        if [ -f "$WORKSPACE_DIR/.config/mise/config.toml" ]; then
            mise trust "$WORKSPACE_DIR/.config/mise/config.toml" >/dev/null 2>&1 || true
        fi
        eval "$(mise activate bash)"

        # Add npm binaries to PATH if they exist
        NPM_PREFIX=$(npm config get prefix 2>/dev/null || echo "")
        if [ -n "$NPM_PREFIX" ]; then
            export PATH="$NPM_PREFIX/bin:$PATH"
        fi

        # Run the command
        if [ $# -eq 0 ]; then
            exec /bin/zsh -l
        else
            exec "$@"
        fi
        ' -- "$@"
    fi
fi
