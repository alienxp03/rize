# Rize zsh defaults
# Loaded from ~/.zshrc to keep Dockerfile clean.

eval "$(
  ~/.local/bin/mise activate zsh \
    | sed -E 's/\x1B\[[0-9;]*[mK]//g' \
    | sed -E '/^mise WARN/d'
)"
HISTFILE=~/.local/share/rize/zsh_history
mkdir -p ~/.local/share/rize

# Homebrew shellenv (full image only)
if [ -x /home/linuxbrew/.linuxbrew/bin/brew ]; then
  eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
fi

alias oc="opencode"
alias ccd="claude --dangerously-skip-permissions"
alias zai='ANTHROPIC_AUTH_TOKEN=$ZAI_API_KEY ANTHROPIC_MODEL=glm-4.7 ANTHROPIC_BASE_URL="https://api.z.ai/api/anthropic" claude --dangerously-skip-permissions'
alias cox="codex --dangerously-bypass-approvals-and-sandbox"
alias gemini="gemini"
