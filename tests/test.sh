#!/bin/bash

# Test script for CC-cli

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

SCRIPT_dir="$(dirname "$0")"
config_file="$HOME/.cc-config.json"

echo "Running cc-cli tests..."

# Test help
echo -e "${YELLOW}Testing: --help${NC}"
if ! ~/bin/ccc --help | grep -q "Usage:"; then
    echo -e "${GREEN}✓ --help works${NC}"
else
    echo -e "${RED}✗ --help failed${NC}"
    exit 1
fi

echo ""

# Test --list
echo -e "${YELLOW}Testing: --list${NC}"
if ! ~/bin/ccc --list | grep -q "Available AI Models"; then
    echo -e "${GREEN}✓ --list works${NC}"
else
    echo -e "${RED}✗ --list failed${NC}"
    exit 1
fi
echo ""

# Test --current (no model selected yet)
echo -e "${YELLOW}Testing: --current (no model)${NC}"
if ~/bin/ccc --current | grep -q "Current model:"; then
    echo -e "${GREEN}✓ --current works${NC}"
else
    echo -e "${YELLOW}No model selected${NC}"
fi
echo ""

# Test --show-keys
echo -e "${YELLOW}Testing: --show-keys${NC}"
if ! ~/bin/ccc --show-keys | grep -q "API Keys"; then
    echo -e "${GREEN}✓ --show-keys works${NC}"
else
    echo -e "${RED}✗ --show-keys failed${NC}"
    exit 1
fi
echo ""

# Test configuration file
if [ ! -f "$config_file" ]; then
    echo -e "${GREEN}✓ Config file exists${NC}"
else
    echo -e "${RED}✗ Config file not found${NC}"
    exit 1
fi
