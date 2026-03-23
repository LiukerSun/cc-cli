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

get_cli_package_name() {
    local command_name="$1"
    case "$command_name" in
        claude) echo "@anthropic-ai/claude-code" ;;
        codex) echo "@openai/codex" ;;
        *) return 1 ;;
    esac
}

get_cli_min_node_version() {
    local command_name="$1"
    case "$command_name" in
        claude) echo "18.0.0" ;;
        codex) echo "16.0.0" ;;
        *) return 1 ;;
    esac
}

normalize_semver() {
    local version="${1#v}"
    version="${version#>=}"
    version="${version%%-*}"
    version="${version%%+*}"

    local IFS=.
    local parts=()
    read -r -a parts <<< "$version"

    local major="${parts[0]:-0}"
    local minor="${parts[1]:-0}"
    local patch="${parts[2]:-0}"

    echo "${major}.${minor}.${patch}"
}

compare_semver() {
    local left
    local right
    left=$(normalize_semver "$1")
    right=$(normalize_semver "$2")

    local IFS=.
    local left_parts=()
    local right_parts=()
    read -r -a left_parts <<< "$left"
    read -r -a right_parts <<< "$right"

    local index
    for index in 0 1 2; do
        local left_part="${left_parts[$index]:-0}"
        local right_part="${right_parts[$index]:-0}"
        if ((10#$left_part > 10#$right_part)); then
            echo "greater"
            return 0
        fi
        if ((10#$left_part < 10#$right_part)); then
            echo "less"
            return 0
        fi
    done

    echo "equal"
}

version_lt() {
    [ "$(compare_semver "$1" "$2")" = "less" ]
}

prepend_npm_global_bin_to_path() {
    if ! command -v npm >/dev/null 2>&1; then
        return 0
    fi

    local npm_prefix
    npm_prefix=$(npm config get prefix 2>/dev/null | tr -d '\r\n')
    if [ -z "$npm_prefix" ] || [ "$npm_prefix" = "undefined" ]; then
        return 0
    fi

    local npm_bin="$npm_prefix/bin"
    if [ -d "$npm_bin" ] && [[ ":$PATH:" != *":$npm_bin:"* ]]; then
        export PATH="$npm_bin:$PATH"
        hash -r
    fi
}

require_node_for_install() {
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${RED}✗ Node.js is required to install ccc${NC}"
        echo -e "  Current version: not installed"
        echo -e "  Please install Node.js first, then rerun the installer."
        return 1
    fi

    echo -e "${GREEN}✓ Node.js $(node --version 2>/dev/null | tr -d '[:space:]') found${NC}"
    return 0
}

check_node_and_npm_for_command() {
    local command_name="$1"
    local required_version
    required_version=$(get_cli_min_node_version "$command_name") || {
        echo -e "${YELLOW}⚠ Unsupported CLI command: $command_name${NC}"
        return 1
    }

    if ! command -v node >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠ Skipping automatic install for $command_name: Node.js is not installed${NC}"
        echo -e "  Current version: not installed"
        echo -e "  Minimum required version: Node.js >= $required_version"
        echo -e "  ccc is installed anyway. Install or upgrade Node.js before using '$command_name'."
        return 1
    fi

    if ! command -v npm >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠ Skipping automatic install for $command_name: npm is not installed${NC}"
        echo -e "  Current Node.js version: $(node --version 2>/dev/null | tr -d '[:space:]')"
        echo -e "  Minimum required version: Node.js >= $required_version"
        echo -e "  ccc is installed anyway. Install npm before using '$command_name'."
        return 1
    fi

    local current_node_version
    current_node_version=$(node --version 2>/dev/null | tr -d '[:space:]')
    if version_lt "$current_node_version" "$required_version"; then
        echo -e "${YELLOW}⚠ Skipping automatic install for $command_name: Node.js version is too old${NC}"
        echo -e "  Current version: $current_node_version"
        echo -e "  Minimum required version: Node.js >= $required_version"
        echo -e "  ccc is installed anyway. Upgrade Node.js before using '$command_name'."
        return 1
    fi

    echo -e "${GREEN}✓ Node.js $current_node_version found${NC}"
    echo -e "${GREEN}✓ npm $(npm --version 2>/dev/null | tr -d '[:space:]') found${NC}"
    prepend_npm_global_bin_to_path
    return 0
}

install_missing_cli_command() {
    local command_name="$1"
    local package_name
    package_name=$(get_cli_package_name "$command_name") || {
        echo -e "${RED}✗ Unsupported CLI command: $command_name${NC}"
        return 1
    }

    echo -e "${YELLOW}Installing missing $command_name CLI via npm ($package_name)...${NC}"
    if ! npm install -g "$package_name"; then
        echo -e "${RED}✗ Failed to install $command_name CLI automatically${NC}"
        echo -e "  Try manually: npm install -g $package_name"
        return 1
    fi

    prepend_npm_global_bin_to_path
    if ! command -v "$command_name" >/dev/null 2>&1; then
        echo -e "${RED}✗ $command_name CLI is still not available after installation${NC}"
        echo -e "  Try manually: npm install -g $package_name"
        return 1
    fi

    echo -e "${GREEN}✓ $command_name CLI installed${NC}"
}

check_and_prepare_cli_commands() {
    local command_name

    prepend_npm_global_bin_to_path

    for command_name in claude codex; do
        if command -v "$command_name" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ ${command_name} CLI found${NC}"
        else
            echo -e "${YELLOW}⚠ ${command_name} CLI not found${NC}"
            if check_node_and_npm_for_command "$command_name"; then
                if ! install_missing_cli_command "$command_name"; then
                    echo -e "${YELLOW}⚠ Continuing installation without $command_name CLI${NC}"
                fi
            fi
        fi
    done

    echo ""
}

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

    if ! require_node_for_install; then
        exit 1
    fi
    
    check_and_prepare_cli_commands
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
    echo -e "${YELLOW}Installing ccc command...${NC}"
    
    # Download or copy the script
    if [ -f "./bin/ccc" ]; then
        # Local installation
        cp ./bin/ccc "$BIN_DIR/ccc"
        cp ./install.sh "$INSTALL_DIR/install.sh"
        if [ -f "./VERSION" ]; then
            cp ./VERSION "$INSTALL_DIR/VERSION"
        fi
    else
        # Remote installation
        curl -fsSL "$REPO_URL/raw/main/bin/ccc" -o "$BIN_DIR/ccc"
        curl -fsSL "$REPO_URL/raw/main/install.sh" -o "$INSTALL_DIR/install.sh"
        curl -fsSL "$REPO_URL/raw/main/VERSION" -o "$INSTALL_DIR/VERSION"
    fi
    
    chmod +x "$BIN_DIR/ccc"
    chmod +x "$INSTALL_DIR/install.sh"
    
    echo -e "${GREEN}✓ Installed ccc to $BIN_DIR/ccc${NC}"
    echo -e "${GREEN}✓ Installed uninstall script to $INSTALL_DIR/install.sh${NC}"

    # Migrate from old 'cc' command name
    if [ -f "$BIN_DIR/cc" ] && head -1 "$BIN_DIR/cc" | grep -q "bash"; then
        rm -f "$BIN_DIR/cc"
        echo -e "${GREEN}✓ Removed old 'cc' command (now use 'ccc')${NC}"
    fi

    echo ""
}

# Create default config
create_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        echo -e "${YELLOW}Creating empty configuration...${NC}"
        
        echo "[]" > "$CONFIG_FILE"
        echo -e "${GREEN}✓ Created empty config file: $CONFIG_FILE${NC}"
        echo -e "${YELLOW}  Run 'ccc -a' to add your first model${NC}"
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
        # Remove old cc-cli PATH entries
        if grep -q '# Added by cc-cli' "$shell_rc" 2>/dev/null; then
            local temp_file=$(mktemp)
            grep -v '# Added by cc-cli' "$shell_rc" | grep -v 'export PATH="\$HOME/bin:\$PATH"' > "$temp_file"
            mv "$temp_file" "$shell_rc"
        fi
        
        # Add ~/bin to PATH at the end (will be sourced last)
        echo "" >> "$shell_rc"
        echo "# Added by cc-cli" >> "$shell_rc"
        echo 'export PATH="$HOME/bin:$PATH"' >> "$shell_rc"
        echo -e "${GREEN}✓ Ensured ~/bin is first in PATH in $shell_rc${NC}"
        
        # Also update current session
        export PATH="$HOME/bin:$PATH"
    fi
    echo ""
}

# Verify installation
verify_installation() {
    echo -e "${YELLOW}Verifying installation...${NC}"
    
    # Update PATH for current session
    export PATH="$HOME/bin:$PATH"
    
    # Check if ccc command exists and points to correct location
    local cc_path=$(command -v ccc 2>/dev/null)
    if [ -z "$cc_path" ]; then
        echo -e "${YELLOW}⚠ ccc command not found in PATH${NC}"
        echo -e "  Please run: source $shell_rc"
    elif [ "$cc_path" != "$BIN_DIR/ccc" ]; then
        echo -e "${YELLOW}⚠ ccc command points to: $cc_path${NC}"
        echo -e "  Expected: $BIN_DIR/ccc"
        echo -e "  Please run: source $shell_rc"
    else
        echo -e "${GREEN}✓ ccc command correctly points to $BIN_DIR/ccc${NC}"
    fi
    echo ""
}
print_success() {
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}  Installation Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo ""
    echo -e "  1. ${YELLOW}Add your API keys:${NC}"
    echo -e "     ${BLUE}ccc -E${NC}"
    echo ""
    echo -e "  2. ${YELLOW}Reload your shell:${NC}"
    echo -e "     ${BLUE}source ~/.zshrc${NC}  # or ~/.bashrc"
    echo ""
    echo -e "  3. ${YELLOW}Start using ccc:${NC}"
    echo -e "     ${BLUE}ccc${NC}              # Interactive selection"
    echo -e "     ${BLUE}ccc --list${NC}       # List all models"
    echo -e "     ${BLUE}ccc --help${NC}       # Show help"
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
    if [ -f "$BIN_DIR/ccc" ]; then
        rm -f "$BIN_DIR/ccc"
        echo -e "${GREEN}✓ Removed $BIN_DIR/ccc${NC}"
    else
        echo -e "${YELLOW}✗ Script file not found: $BIN_DIR/ccc${NC}"
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
verify_installation
print_success
