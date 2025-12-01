# Dockerize - Multi-Language AI Agent Container

A flexible Docker container for running AI agents (claude-code) with multi-language support via mise version manager.

## Features

- **Multi-language support with mise**: Go, Node.js, Ruby, Python with multiple configurable versions
- **Version management**: Specify versions at build time (e.g., `--go=1.23.3,1.22.5`)
- **claude-code ready**: All setup needed for Anthropic's claude-code AI agent
- **Beautiful shell**: Powerline10k theme with fzf integration (from Claude Code's devcontainer)
- **Modern tooling**: Git, ripgrep, fd, bat, jq, zsh included
- **Simple wrapper script**: One-command interface for building and running containers
- **Lazy installation**: Tools are downloaded on first use (fast Docker builds)

## Quick Start

### 1. Make the Wrapper Script Executable

```bash
chmod +x ./dockerize
```

### 2. Add to Your PATH

Copy the `dockerize` script to a directory in your PATH, or add the current directory to PATH:

```bash
# Option A: Copy to /usr/local/bin
sudo cp dockerize /usr/local/bin/

# Option B: Add current directory to PATH
export PATH="$(pwd):$PATH"
```

### 3. Build and Start Using

```bash
# Build with default versions (Go 1.23.3, Node 24)
dockerize build

# Or build with custom versions
dockerize build --go=1.23.3,1.22.5 --node=24,22

# Start interactive shell in current directory
dockerize shell

# Run a command in the container
dockerize run claude-code .
dockerize exec go version
dockerize exec node --version
```

## Usage Examples

### Interactive Shell

```bash
cd ~/my-go-project
dockerize shell
# Now inside container with tools available
```

### Run Claude Code

```bash
cd ~/my-project
dockerize run claude-code .
```

### Check Installed Languages

```bash
dockerize exec go version
dockerize exec node --version
```

### Build with Specific Versions

```bash
# Multiple Go versions
dockerize build --go=1.23.3,1.22.5,1.21.0

# Multiple versions of multiple tools
dockerize build --go=1.23.3,1.22.5 --node=24,22
```

## Shell & Terminal

The container uses **zsh with Powerline10k** theme and **fzf** (fuzzy finder), matching the setup from Claude Code's own devcontainer. This gives you:

- ✨ Beautiful Powerline10k prompt with git status
- 🔍 Fuzzy file finder (fzf) integration
- 📝 Syntax highlighting and auto-completion
- No configuration prompts - it just works!

The shell is automatically initialized when you start a container, so you can start working immediately.

## Supported Languages

The container supports the following languages via mise:

- **Go**: Default 1.23.3 (configurable)
- **Node.js**: Default 24 (configurable)
- **Ruby**: Available (requires build tools - see customization)
- **Python**: Available (requires build tools - see customization)

Go and Node.js have prebuilt binaries and install quickly. Ruby and Python require build tools and compilation, which is why they're optional.

## Installed Tools

### Global NPM Packages

- **@anthropic-ai/claude-code**: AI-powered code assistant
- **codex**: OpenAI Codex command-line interface
- **opencode**: OpenCode AI assistant

Access them directly:

```bash
dockerize claude-code --help
dockerize codex --help
dockerize opencode --help
```

### System Utilities

- **git**: Version control
- **ripgrep**: Fast file search
- **fd**: Find replacement
- **bat**: Enhanced cat with syntax highlighting
- **jq**: JSON processor
- **zsh**: Modern shell

## Troubleshooting

### Permission Denied Errors

If you see permission errors when editing files:

```
Error: EACCES: permission denied
```

This typically means the user ID mapping isn't working. Ensure your alias includes:

```bash
--user $(id -u):$(id -g)
```

### Docker Daemon Not Running

```bash
# macOS
open /Applications/Docker.app

# Linux
sudo systemctl start docker
```

### Port Already in Use

If you're running services inside the container that conflict:

```bash
# Add port mappings to the alias
dockerize -p 3000:3000 -p 8080:8080
```

### Verify Installation

```bash
dockerize bash -c 'echo "=== Versions ===" && \
  go version && \
  ruby --version && \
  node --version && \
  python3 --version && \
  echo "" && \
  echo "=== AI Tools ===" && \
  which claude-code'
```

### First Build Takes Time

The first Docker build downloads and installs all dependencies (~1-2 GB). This typically takes 5-10 minutes depending on your internet connection. Subsequent builds are much faster.

## Advanced Usage

### Custom Working Directory

```bash
# Mount a different directory
docker run --rm -it \
  --user $(id -u):$(id -g) \
  -v /path/to/project:/workspace \
  dockerize:latest
```

### Run Multiple Commands

```bash
dockerize bash -c "go build && ruby script.rb && claude-code ."
```

### Port Mapping

```bash
# For local services (web servers, APIs, etc.)
docker run --rm -it \
  --user $(id -u):$(id -g) \
  -v "$(pwd):/workspace" \
  -p 3000:3000 \
  -p 8080:8080 \
  dockerize:latest
```

## Building from Dockerfile

If you want to customize the image:

1. Modify `Dockerfile` as needed
2. Rebuild:
   ```bash
   docker build -t dockerize:latest .
   ```
3. The alias will automatically use the new image

### Common Customizations

**Add more system packages:**

```dockerfile
RUN apt-get install -y postgresql-client
```

**Add more global npm packages:**

```dockerfile
RUN npm install -g some-tool
```

**Change default shell:**

```dockerfile
ENV SHELL=/bin/bash
CMD ["/bin/bash"]
```

## System Requirements

- **Docker**: 20.10+ recommended
- **Disk space**: 2-3 GB for the image and language runtimes
- **Memory**: 2GB minimum (4GB+ recommended for large projects)

## Customization

To customize the image:

- Add more system packages via apt-get in the Dockerfile
- Add more global npm packages
- Change the versions of Go, Ruby, etc.
- Use it as a base for your own specialized containers

## Useful Links

- [Claude Code](https://github.com/anthropics/claude-code)

## License

MIT
