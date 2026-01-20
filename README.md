# Rize

Containerized development environment for AI coding agents. Pre-configured with multiple language runtimes, AI tools, and a full development toolbelt.

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/alienxp03/rize/refs/heads/master/rize | bash
```

This command installs the `rize` CLI to `/usr/local/bin` and pulls the pre-built Docker image (~5GB).

**Requirements:** Docker

---

## What's Installed

### AI Coding Agents

| Agent | Command | Description |
|-------|---------|-------------|
| [Claude Code](https://github.com/anthropics/claude-code) | `rize claude` | Anthropic's AI coding assistant |
| [Codex](https://github.com/openai/codex) | `rize codex` | OpenAI's code generation agent |
| [OpenCode](https://github.com/opencode-ai/opencode) | `rize opencode` | Multi-model AI coding tool |
| Gemini | `rize gemini` | Google's Gemini (via OpenCode) |

### Language Runtimes (via [mise](https://mise.jdx.dev))

| Language | Version |
|----------|---------|
| Node.js | 24 |
| Python | 3.12 |
| Ruby | 3.4 |
| Go | 1.25 |
| Rust | 1.92 |

### Package Managers & Build Tools

- **Node:** npm, yarn, pnpm
- **Python:** pip, pipx, uv
- **Ruby:** bundler
- **Go:** go modules
- **Rust:** cargo

### Developer Tools

| Category | Tools |
|----------|-------|
| Shell | zsh + Oh My Zsh + Powerlevel10k |
| Editors | vim, nano |
| Search | ripgrep (`rg`), fd, jq |
| Git | git, git-lfs |
| Databases | PostgreSQL 18 (client + server), Redis 8, SQLite3 |
| Network | curl, wget, ping, dig, netstat |
| Utils | htop, lsof, less, tar, zip, rsync |
| Package Manager | Homebrew (full image only) |

### Code Quality Tools

- **JavaScript:** eslint, prettier
- **Go:** golangci-lint, delve (debugger)
- **Python:** yq
- **Ruby:** (use bundler to install rubocop per-project)

---

## Usage

### Run AI Agents

```bash
rize claude              # Start Claude Code
rize codex               # Start Codex
rize opencode            # Start OpenCode
rize gemini              # Start Gemini (via OpenCode)
```

All agents run with permissions auto-approved since they're sandboxed in the container.

### Shell Access

```bash
rize shell               # Interactive zsh shell
rize exec <command>      # Run a single command
```

### Examples

```bash
# Start Claude in current project
cd ~/projects/my-app
rize claude

# Run tests
rize exec npm test

# Install dependencies
rize exec bundle install

# Interactive debugging
rize shell
```

---

## Configuration

### API Keys

Create `~/.env` or `~/.rize/env` on your host:

```bash
# AI Agent API Keys
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=...

# Optional: Language-specific config
GOPROXY=https://proxy.company.com,direct
GOPRIVATE=gitlab.company.com/*
```

These files are automatically loaded when the container starts.

### Git & SSH

The following are auto-mounted from your host (read-only):

- `~/.gitconfig` - Git configuration
- `~/.netrc` - HTTP authentication for git/curl
- `~/.ssh/known_hosts` - SSH known hosts
- `SSH_AUTH_SOCK` - SSH agent forwarding

### Environment Variables

| Variable | Description |
|----------|-------------|
| `RIZE_IMAGE` | Override Docker image (e.g., `alienxp03/rize:2026-01-15`) |
| `RIZE_GEMINI_MODEL` | Gemini model for `rize gemini` (default: `gemini-pro`) |
| `RIZE_WORKSPACE_UNIQUE` | Set to `1` to add path hash to workspace dir |

---

## How It Works

### Workspace Mounting

Your current directory is mounted to `/workspace/<project-name>`:

```
~/projects/my-app  →  /workspace/my-app
~/work/my-app      →  /workspace/my-app-a1b2c3d4  (with RIZE_WORKSPACE_UNIQUE=1)
```

### Persistent Data

| Data | Location |
|------|----------|
| Shell history | `~/.rize/zsh_history` |
| Agent configs | Docker volume `rize-agents` |
| Claude settings | Shared from `~/.claude/` |

### Docker Socket

The Docker socket is mounted, allowing you to run Docker commands inside the container (Docker-outside-of-Docker).

---

## Image Variants

### Full (default)

```bash
docker pull alienxp03/rize:latest
```

Includes everything: all runtimes, Homebrew, AI agents, database servers.

### Slim

```bash
docker build -t rize:slim --target slim .
```

Minimal image with runtimes only. No Homebrew or database servers.

---

## Uninstall

```bash
# Remove CLI
rm -f /usr/local/bin/rize

# Remove Docker resources
docker image rm alienxp03/rize:latest
docker volume rm rize-agents
docker volume rm $(docker volume ls -q | grep rize-vendor) 2>/dev/null || true

# Remove config (optional)
rm -rf ~/.rize
```

---

## Development

```bash
# Build locally
make build

# Test
make exec echo "hello"

# Install from local source
make install
```
