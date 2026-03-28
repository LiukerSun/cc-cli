#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'chmod -R u+w "$TMP_DIR" 2>/dev/null || true; rm -rf "$TMP_DIR"' EXIT

HOME_DIR="$TMP_DIR/home"
BIN_DIR="$HOME_DIR/.local/bin"
CONFIG_FILE="$HOME_DIR/.config/ccc/config.json"
CCC_BIN="$BIN_DIR/ccc"
GO_BIN_DIR="$(dirname "$(command -v go)")"

mkdir -p "$BIN_DIR"

export HOME="$HOME_DIR"
export PATH="$BIN_DIR:$GO_BIN_DIR:/usr/bin:/bin"
export GOCACHE="$TMP_DIR/go-cache"
export GOMODCACHE="$TMP_DIR/go-mod-cache"
export XDG_CONFIG_HOME="$HOME_DIR/.config"
export XDG_DATA_HOME="$HOME_DIR/.local/share"
export XDG_CACHE_HOME="$HOME_DIR/.cache"
export XDG_STATE_HOME="$HOME_DIR/.local/state"

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"

    if [[ "$haystack" != *"$needle"* ]]; then
        echo "$message" >&2
        exit 1
    fi
}

(
    cd "$REPO_DIR"
    go build -o "$CCC_BIN" ./cmd/ccc
)

help_output="$("$CCC_BIN" help)"
assert_contains "$help_output" "ccc profile add" "expected help output to include profile commands"

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

profile_list_output="$("$CCC_BIN" profile list)"
assert_contains "$profile_list_output" "claude-test" "expected profile list to include generated profile id"

current_output="$("$CCC_BIN" current)"
assert_contains "$current_output" "Name: Claude Test" "expected current profile output to include Claude Test"

if ! "$CCC_BIN" profile update claude-test \
    --name "Claude Prod" \
    --id claude-prod \
    --env FOO=bar \
    --no-sync > "$TMP_DIR/profile-update.txt"; then
    echo "expected profile update to succeed" >&2
    cat "$TMP_DIR/profile-update.txt" >&2 || true
    exit 1
fi

updated_current_output="$("$CCC_BIN" current)"
assert_contains "$updated_current_output" "ID: claude-prod" "expected current profile output to include updated id"

dry_run_output="$("$CCC_BIN" run --dry-run)"
assert_contains "$dry_run_output" "Command: claude" "expected dry-run output to target claude"

config_show_output="$("$CCC_BIN" config show)"
assert_contains "$config_show_output" '"current_profile": "claude-prod"' "expected config show output to include current profile"

assert_contains "$config_show_output" '"FOO": "bar"' "expected config show output to include updated env"

upgrade_output="$("$CCC_BIN" upgrade --version 2.2.1 --dry-run)"
assert_contains "$upgrade_output" "Target version: 2.2.1" "expected upgrade --dry-run output to include target version"

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
