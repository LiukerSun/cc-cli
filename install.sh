#!/bin/bash

# CC-CLI Installation Script
# https://github.com/LiukerSun/cc-cli

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0;0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION_FILE="$SCRIPT_DIR/VERSION"
if [ -f "$VERSION_FILE" ]; then
    VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
else
    VERSION="unknown"
fi

REPO_URL="https://github.com/LiukerSun/cc-cli"

# Installation paths
INSTALL_DIR="$HOME/.cc-cli"
BIN_DIR="$HOME/bin"
CONFIG_FILE="$HOME/.cc-config.json"

echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}  CC-CLI Installer v${VERSION}${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo ""

# Check system requirements
check_requirements() {
    echo -e "${YELLOW}Checking system requirements...${NC}"
    
    # Check for bash
    if [ -z "$BASH_VERSION" ]; then
        echo -e "${RED}✗ Bash is required but not found${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Bash ${BASH_VERSION%%(*)} found${NC}"
    
    # Check for claude command
    if ! command -v claude &> /dev/null; then
        echo -e "${YELLOW}⚠ Claude CLI not found (optional)${NC}"
        echo -e "  Install from: https://claude.ai"
    else
        echo -e "${GREEN}✓ Claude CLI found${NC}"
    fi
    
    echo ""
}

# Create directories
create_directories() {
    echo -e "${YELLOW}Creating directories...${NC}"
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$BIN_DIR"
    
    echo -e "${GREEN}✓ Created $INSTALL_DIR${NC}"
    echo -e "${GREEN}✓ Created $BIN_DIR${NC}"
    echo ""
}

# Install main script
install_script() {
    echo -e "${YELLOW}Installing cc command...${NC}"
    
    # Download or copy the script
    if [ -f "./bin/cc" ]; then
        # Local installation
        cp ./bin/cc "$BIN_DIR/cc"
    else
        # Remote installation
        curl -fsSL "$REPO_URL/raw/main/bin/cc" -o "$BIN_DIR/cc"
    fi
    
    chmod +x "$BIN_DIR/cc"
    
    echo -e "${GREEN}✓ Installed cc to $BIN_DIR/cc${NC}"
    echo ""
}

# Create default config
create_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        echo -e "${YELLOW}Creating empty configuration...${NC}"
        
        echo "[]" > "$CONFIG_FILE"
        echo -e "${GREEN}✓ Created empty config file: $CONFIG_FILE${NC}"
        echo -e "${YELLOW}  Run 'cc -a' to add your first model${NC}"
    else
        echo -e "${GREEN}✓ Config file already exists: $CONFIG_FILE${NC}"
    fi
    echo ""
}

# Update shell configuration
update_shell() {
    echo -e "${YELLOW}Configuring shell...${NC}"
    
    local shell_rc=""
    if [ -f "$HOME/.zshrc" ]; then
        shell_rc="$HOME/.zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        shell_rc="$HOME/.bashrc"
    fi
    
    if [ -n "$shell_rc" ]; then
        # Check if PATH already includes ~/bin
        if ! grep -q 'export PATH="$HOME/bin:$PATH"' "$shell_rc" 2>/dev/null; then
            echo "" >> "$shell_rc"
            echo "# Added by cc-cli" >> "$shell_rc"
            echo 'export PATH="$HOME/bin:$PATH"' >> "$shell_rc"
            echo -e "${GREEN}✓ Added ~/bin to PATH in $shell_rc${NC}"
        else
            echo -e "${GREEN}✓ ~/bin already in PATH${NC}"
        fi
    fi
    echo ""
}

# Print success message
print_success() {
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}  Installation Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo ""
    echo -e "  1. ${YELLOW}Add your API keys:${NC}"
    echo -e "     ${BLUE}cc -E${NC}"
    echo ""
    echo -e "  2. ${YELLOW}Reload your shell:${NC}"
    echo -e "     ${BLUE}source ~/.zshrc${NC}  # or ~/.bashrc"
    echo ""
    echo -e "  3. ${YELLOW}Start using cc:${NC}"
    echo -e "     ${BLUE}cc${NC}              # Interactive selection"
    echo -e "     ${BLUE}cc --list${NC}       # List all models"
    echo -e "     ${BLUE}cc --help${NC}       # Show help"
    echo ""
    echo -e "${BLUE}Documentation:${NC}"
    echo -e "  $REPO_URL"
    echo ""
    echo -e "${BLUE}Configuration file:${NC}"
    echo -e "  $CONFIG_FILE"
    echo ""
}

# Uninstall function
uninstall() {
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  CC-CLI Uninstaller${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo ""
    
    read -p "Are you sure you want to uninstall cc-cli? (y/N): " confirm
    if [[ ! "$confirm" =~ ^[Yy](es)?$ ]]; then
        echo "Uninstall cancelled."
        exit 0
    fi
    
    echo ""
    echo -e "${YELLOW}Removing files...${NC}"
    
    # Remove main script
    if [ -f "$BIN_DIR/cc" ]; then
        rm -f "$BIN_DIR/cc"
        echo -e "${GREEN}✓ Removed $BIN_DIR/cc${NC}"
    else
        echo -e "${YELLOW}✗ Script file not found: $BIN_DIR/cc${NC}"
    fi
    
    # Remove installation directory
    if [ -d "$INSTALL_DIR" ]; then
        rm -rf "$INSTALL_DIR"
        echo -e "${GREEN}✓ Removed $INSTALL_DIR${NC}"
    else
        echo -e "${YELLOW}✗ Installation directory not found: $INSTALL_DIR${NC}"
    fi
    
    # Remove temp env file
    ENV_FILE="/tmp/cc-model-env.sh"
    if [ -f "$ENV_FILE" ]; then
        rm -f "$ENV_FILE"
        echo -e "${GREEN}✓ Removed $ENV_FILE${NC}"
    fi
    
    # Remove config file (optional)
    echo ""
    read -p "Delete config file? (y/N): " config_confirm
    if [[ "$config_confirm" =~ ^[Yy](es)?$ ]]; then
        if [ -f "$CONFIG_FILE" ]; then
            rm -f "$CONFIG_FILE"
            echo -e "${GREEN}✓ Removed $CONFIG_FILE${NC}"
        else
            echo -e "${YELLOW}✗ Config file not found: $CONFIG_FILE${NC}"
        fi
    else
        echo -e "${YELLOW}✓ Config file preserved: $CONFIG_FILE${NC}"
    fi
    
    # Remove Claude settings (optional)
    CLAUDE_SETTINGS_FILE="$HOME/.claude/settings.json"
    if [ -f "$CLAUDE_SETTINGS_FILE" ]; then
        echo ""
        read -p "Remove cc-cli entries from Claude settings? (y/N): " settings_confirm
        if [[ "$settings_confirm" =~ ^[Yy](es)?$ ]]; then
            if command -v jq &> /dev/null; then
                # Use jq to clean settings
                jq 'del(.env.ANTHROPIC_MODEL, .env.ANTHROPIC_SMALL_FAST_MODEL, .env.CLAUDE_CODE_MODEL, .env.CLAUDE_CODE_SMALL_MODEL, .env.CLAUDE_CODE_SUBAGENT_MODEL, .model) | .permissions.deny = (.permissions.deny // [] | map(select(. != "Agent(Explore)"))) | if (.permissions.deny | length) == 0 then del(.permissions.deny) else . end | if (.env | length) == 0 then del(.env) else . end | if (.permissions | length) == 0 then del(.permissions) else . end' "$CLAUDE_SETTINGS_FILE" > "$CLAUDE_SETTINGS_FILE.tmp" && mv "$CLAUDE_SETTINGS_FILE.tmp" "$CLAUDE_SETTINGS_FILE"
                echo -e "${GREEN}✓ Cleaned $CLAUDE_SETTINGS_FILE${NC}"
            else
                echo -e "${YELLOW}✗ jq not found, skipping settings cleanup${NC}"
            fi
        else
            echo -e "${YELLOW}✓ Claude settings preserved: $CLAUDE_SETTINGS_FILE${NC}"
        fi
    fi
    
    # Remove from shell configuration
    echo ""
    echo -e "${YELLOW}Cleaning shell configuration...${NC}"
    
    for shell_rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
        if [ -f "$shell_rc" ]; then
            if grep -q '# Added by cc-cli' "$shell_rc" 2>/dev/null; then
                # Remove cc-cli entries
                local temp_file=$(mktemp)
                awk '
                    /^# Added by cc-cli$/ { skip=1; next }
                    /^export PATH="\$HOME\/bin:\$PATH"$/ { if (skip) next }
                    skip && /^$/ { next }
                    skip && !/^export PATH/ { skip=0 }
                    { print }
                ' "$shell_rc" > "$temp_file" && mv "$temp_file" "$shell_rc"
                echo -e "${GREEN}✓ Cleaned $shell_rc${NC}"
            else
                echo -e "${YELLOW}✗ No cc-cli entries in $shell_rc${NC}"
            fi
        fi
    done
    
    # Reload shell config reminder
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}  Uninstall Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo -e "${YELLOW}  Please restart your shell or run:${NC}"
    echo -e "  ${BLUE}source ~/.zshrc${NC}  # or ~/.bashrc"
    echo ""
    
    exit 0
}

# Parse arguments
case "${1:-}" in
    --uninstall)
        uninstall
        ;;
    --help|-h)
        echo "Usage: ./install.sh [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --uninstall    Remove cc-cli"
        echo "  --help, -h     Show this help message"
        exit 0
        ;;
esac

# Run installation
check_requirements
create_directories
install_script
create_config
update_shell
print_success
