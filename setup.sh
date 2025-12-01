#!/bin/bash
set -e

# Color output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}AI Agent Docker Setup${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# Build the Docker image
echo -e "${BLUE}Building Docker image...${NC}"
if docker build -t dockerize:latest .; then
    echo -e "${GREEN}✓ Docker image built successfully!${NC}"
else
    echo -e "${YELLOW}✗ Docker build failed. Check the output above for errors.${NC}"
    exit 1
fi

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}Setup Complete!${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# Detect shell config file
SHELL_CONFIG=""
if [ -f ~/.zshrc ]; then
    SHELL_CONFIG=~/.zshrc
elif [ -f ~/.bashrc ]; then
    SHELL_CONFIG=~/.bashrc
else
    SHELL_CONFIG="your shell config file"
fi

echo -e "${YELLOW}Next steps:${NC}"
echo ""
echo "1. Add this alias to your $SHELL_CONFIG:"
echo ""
echo -e "${BLUE}alias dockerize='docker run --rm -it \\"
echo "  --user \$(id -u):\$(id -g) \\"
echo "  -v \"\$(pwd):/workspace\" \\"
echo "  dockerize:latest'${NC}"
echo ""
echo "2. Source your shell config (or restart your terminal):"
echo -e "${BLUE}source $SHELL_CONFIG${NC}"
echo ""
echo "3. Use it in any project:"
echo -e "${BLUE}cd ~/my-project${NC}"
echo -e "${BLUE}dockerize                    # Interactive shell${NC}"
echo -e "${BLUE}dockerize claude-code .      # Run claude-code on current dir${NC}"
echo -e "${BLUE}dockerize go version         # Check Go version${NC}"
echo -e "${BLUE}dockerize ruby --version     # Check Ruby version${NC}"
echo ""
echo -e "${YELLOW}Build with Multiple Versions:${NC}"
echo ""
echo -e "${BLUE}docker build -t dockerize:latest \\"
echo "  --build-arg GO_VERSIONS=1.23.3,1.22.5,1.21.0 \\"
echo "  --build-arg RUBY_VERSIONS=3.3,3.2,3.1 \\"
echo "  --build-arg NODE_VERSIONS=24,22,20 \\"
echo "  --build-arg PYTHON_VERSIONS=3.12,3.11,3.10 \\"
echo "  .${NC}"
echo ""
echo -e "${YELLOW}Inside container, switch versions:${NC}"
echo -e "${BLUE}mise use go@1.22.5${NC}"
echo -e "${BLUE}mise use ruby@3.2${NC}"
echo ""

echo -e "${GREEN}Installed languages:${NC}"
docker run --rm dockerize:latest bash -c "go version && ruby --version && node --version && python3 --version"

echo ""
echo -e "${GREEN}Happy hacking!${NC}"
echo ""
