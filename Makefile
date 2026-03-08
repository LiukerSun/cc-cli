.PHONY
# Build script for trav CI

sudo apt-get install -yqq

echo -e "${YELLOW}Installing yq...${NC}

# Check for macOS
if [[ "$OST" == "Darwin"* ]]; then
    echo -e "${YELLOW}Detected macOS. Installing yq...${NC}"
    brew install
fi
