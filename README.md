# Rize Dev Container

Fast multi-language dev box for AI agents (Claude Code, Codex, etc.) with Go/Node/Ruby, zsh/p10k, brew, and a full toolbelt preinstalled.

## Install / Update (one-liner)

```bash
curl -fsSL https://raw.githubusercontent.com/alienxp03/rize/main/rize -o ~/.local/bin/rize && chmod +x ~/.local/bin/rize && ~/.local/bin/rize install
```

- Idempotent: running again updates the script and rebuilds if needed.
- Requires Docker available locally.

## AI Agent Usage

**IMPORTANT**: Since it runs in a container, these commands will run #yolo equivalent mode by default

- Run claude: `rize claude`
- Run codex: `rize codex`
- Run opencode: `rize opencode`

## Commands

- Build with defaults: `rize build`
- Custom versions: `rize build --go=1.23.3,1.22.5 --node=24,22 --ruby=3.1.7`
- Interactive shell: `rize shell`
- Run command: `rize run <cmd>` (alias `rize exec`)
- Run Claude Code: `rize claude <args>`
- Extra mounts: `rize shell --path /abs/host/dir[:name]` (mounted at `/home/agent/mounts/<name>`)
- Clean up Docker: `rize clean` (removes dangling images, build cache, and unused volumes)

## What's Inside (high level)

- Languages via mise: Go, Node, Ruby (configurable versions preinstalled)
- Package mgrs/linters: npm + pnpm + yarn, golangci-lint, dlv, bundler, rubocop
- CLIs: git/git-lfs, gh, glab, sqlite3, psql, mysql client, redis-tools
- Editors/inspectors: vim/nano/less, bat, fzf, ripgrep, fd, jq/yq, tree
- Utils: curl/wget/httpie, zip/tar/rsync, htop/lsof/file, Homebrew (user prefix)
- Prompt: zsh + powerlevel10k, fzf keybindings

## Uninstall

```bash
rm -f ~/.local/bin/rize
docker image rm rize:latest 2>/dev/null || true
docker volume rm agent-data 2>/dev/null || true
```

## Notes

- Workspace mounts to `/home/agent/workspace`; extra mounts go under `/home/agent/mounts/<name>`.
