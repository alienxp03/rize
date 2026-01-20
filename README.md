# Rize Dev Container

Fast multi-language dev box for AI agents (Claude Code, Codex, etc.) with Go/Node/Ruby, zsh/p10k, brew, and a full toolbelt preinstalled.

## Install / Update (one-liner)

```bash
curl -fsSL https://raw.githubusercontent.com/alienxp03/rize/refs/heads/master/rize | bash
```

- Idempotent: running again updates the script.
- Pulls pre-built Docker image (~5GB) from Docker Hub - no local build required.
- Requires Docker available locally.

## AI Agent Usage

**IMPORTANT**: Since it runs in a container, these commands will run #yolo equivalent mode by default

- Run claude: `rize claude`
- Run codex: `rize codex`
- Run opencode: `rize opencode`

## Commands

- Interactive shell: `rize shell`
- Run command: `rize run <cmd>` (alias `rize exec`)
- Run Claude Code: `rize claude <args>`
- Extra mounts: `rize shell --path /abs/host/dir[:name]` (mounted at `/home/agent/mounts/<name>`)
- Clean up Docker: `rize clean` (removes dangling images, build cache, and unused volumes)

### Multi-Project Support

Each project gets its own workspace directory inside the container:
- `/workspace/my-app` - Project named "my-app"
- `/workspace/my-app-2` - Second project also named "my-app"

The project name is derived from the basename of the current directory. If you have multiple directories with the same name, they automatically get numeric suffixes to avoid conflicts.

## Environment Variables

Create a `~/.env` file on your host to automatically load environment variables in the container:

```bash
# Example: Go proxy configuration for internal GitLab access
GOPROXY=https://proxy.company.com,direct
GOPRIVATE=gitlab.company.com/*,*.company.local
GONOPROXY=gitlab.company.com/*
GONOSUMDB=gitlab.company.com/*

# GitLab authentication
GITLAB_TOKEN=glpat-xxxxxxxxxxxxx
```

The `.env` file is automatically sourced when the container starts, making variables available to all tools (Go, Node, Ruby, Python, etc.). The file is mounted read-only for safety—to update variables, edit `~/.env` on your host and restart the container.

**Note**: This is for global environment variables. Project-specific `.env` files can also be created in the workspace directory and will be respected.

### Rize Overrides

- `RIZE_IMAGE` - Override the image tag or digest (e.g., `alienxp03/rize:2026-01-15` or `alienxp03/rize@sha256:...`) for deterministic runs.
- `RIZE_GEMINI_MODEL` - Override the default Gemini model used by `rize gemini` (defaults to `gemini-pro`).

### .netrc for Authentication

If you have a `~/.netrc` file, it will be automatically mounted and available to tools like `git`, `curl`, and other utilities that support `.netrc` authentication:

```bash
# Example ~/.netrc for GitLab/GitHub authentication
machine github.com
login your-github-username
password ghp_xxxxxxxxxxxxxx

machine gitlab.company.com
login your-gitlab-username
password glpat-xxxxxxxxxxxxxx
```

Make sure `.netrc` has the correct permissions on your host:

```bash
chmod 600 ~/.netrc
```

The file is mounted read-only in the container, and tools will use it for authentication without modification.

## What's Inside (high level)

- Languages via mise: Go, Node, Ruby, Python (configurable versions preinstalled)
- Package mgrs/linters: npm + pnpm + yarn, golangci-lint, dlv, bundler, rubocop
- CLIs: git/git-lfs, gh, glab, sqlite3, psql, mysql client, redis-tools
- Editors/inspectors: vim/nano/less, bat, fzf, ripgrep, fd, jq/yq, tree
- Utils: curl/wget/httpie, zip/tar/rsync, htop/lsof/file, Homebrew (user prefix)
- Prompt: zsh + powerlevel10k, fzf keybindings

## Per-Project Vendor Caching

Rize manages gem/dependency caches on a per-project basis using Docker volumes. Each project gets its own vendor volume:

- Volume naming: `rize-vendor-{project-name}-{path-hash}`
- Example: `rize-vendor-my-app-a1b2c3d4`

This means:

- First `bundle install` in a project compiles gems for Linux and caches them
- Subsequent runs reuse cached gems (fast)
- Switching between projects with different dependencies doesn't cause conflicts

To list all vendor volumes:

```bash
docker volume ls | grep rize-vendor
```

To remove a specific project's vendor cache:

```bash
docker volume rm rize-vendor-project-name-abc12345
```

## Uninstall

```bash
rm -f ~/.local/bin/rize
docker image rm rize:latest 2>/dev/null || true
docker volume rm agent-data 2>/dev/null || true
# Remove project-specific vendor volumes if desired:
docker volume rm $(docker volume ls -q | grep rize-vendor) 2>/dev/null || true
```

## Notes

- Workspace mounts to `/workspace/<project-name>`; extra mounts go under `/home/agent/mounts/<name>`.
- Vendor directory is managed via Docker volumes to avoid filesystem permission conflicts between macOS and Linux.

## Local Builds (Full vs Slim)

If you build locally, you can choose between a full or slim image:

```bash
# Full (default)
docker build -t rize:full --target final .

# Slim (no Homebrew and fewer extra tools)
docker build -t rize:slim --target slim .
```
