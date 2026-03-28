#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

HOME_DIR="$TMP_DIR/home"
BIN_DIR="$HOME_DIR/.local/bin"
CONFIG_FILE="$HOME_DIR/.config/ccc/config.json"
CCC_BIN="$BIN_DIR/ccc"

mkdir -p "$BIN_DIR"

(
    cd "$REPO_DIR"
    HOME="$HOME_DIR" \
    GOCACHE="$TMP_DIR/go-cache" \
    go build -o "$CCC_BIN" ./cmd/ccc
)

HOME="$HOME_DIR"
PATH="$BIN_DIR:/usr/bin:/bin"

if ! "$CCC_BIN" help | grep -q "ccc profile add"; then
    echo "expected help output to include profile commands" >&2
    exit 1
fi

if ! "$CCC_BIN" profile add \
    --preset anthropic \
    --name "Claude Test" \
    --api-key test-key \
    --model test-model > "$TMP_DIR/profile-add.txt"; then
    echo "expected profile add to succeed" >&2
    cat "$TMP_DIR/profile-add.txt" >&2 || true
    exit 1
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "expected config file to be created at $CONFIG_FILE" >&2
    exit 1
fi

if ! "$CCC_BIN" profile list | grep -q "claude-test"; then
    echo "expected profile list to include generated profile id" >&2
    exit 1
fi

if ! "$CCC_BIN" current | grep -q "Name: Claude Test"; then
    echo "expected current profile output to include Claude Test" >&2
    exit 1
fi

if ! "$CCC_BIN" profile update claude-test \
    --name "Claude Prod" \
    --id claude-prod \
    --env FOO=bar \
    --no-sync > "$TMP_DIR/profile-update.txt"; then
    echo "expected profile update to succeed" >&2
    cat "$TMP_DIR/profile-update.txt" >&2 || true
    exit 1
fi

if ! "$CCC_BIN" current | grep -q "ID: claude-prod"; then
    echo "expected current profile output to include updated id" >&2
    exit 1
fi

if ! "$CCC_BIN" run --dry-run | grep -q "Command: claude"; then
    echo "expected dry-run output to target claude" >&2
    exit 1
fi

if ! "$CCC_BIN" config show | grep -q '"current_profile": "claude-prod"'; then
    echo "expected config show output to include current profile" >&2
    exit 1
fi

if ! "$CCC_BIN" config show | grep -q '"FOO": "bar"'; then
    echo "expected config show output to include updated env" >&2
    exit 1
fi

if ! "$CCC_BIN" upgrade --version 2.2.1 --dry-run | grep -q "Target version: 2.2.1"; then
    echo "expected upgrade --dry-run output to include target version" >&2
    exit 1
fi

if ! HOME="$HOME_DIR" bash "$REPO_DIR/bin/ccc" help > "$TMP_DIR/wrapper-stdout.txt" 2> "$TMP_DIR/wrapper-stderr.txt"; then
    echo "expected legacy shell wrapper to delegate to installed Go binary" >&2
    cat "$TMP_DIR/wrapper-stderr.txt" >&2 || true
    exit 1
fi

if ! grep -q "ccc profile add" "$TMP_DIR/wrapper-stdout.txt"; then
    echo "expected wrapper output to come from Go CLI help" >&2
    exit 1
fi

if ! grep -q "legacy compatibility wrapper" "$TMP_DIR/wrapper-stderr.txt"; then
    echo "expected wrapper to print a deprecation warning" >&2
    exit 1
fi

echo "test.sh: ok"
