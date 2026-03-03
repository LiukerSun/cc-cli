#!/bin/bash

CC_SCRIPT="./bin/cc"
CONFIG_FILE="./test-config.json"
PASS=0
FAIL=0

setup() {
    echo '[{"name":"Test Model 1","env":{"ANTHROPIC_BASE_URL":"https://api.test.com","ANTHROPIC_AUTH_TOKEN":"test-key-1","ANTHROPIC_MODEL":"test-1","ANTHROPIC_SMALL_FAST_MODEL":"test-1-small"}},{"name":"Test Model 2","env":{"ANTHROPIC_BASE_URL":"https://api.test2.com","ANTHROPIC_AUTH_TOKEN":"test-key-2","ANTHROPIC_MODEL":"test-2","ANTHROPIC_SMALL_FAST_MODEL":"test-2-small"}}]' > "$CONFIG_FILE"
}

cleanup() {
    rm -f "$CONFIG_FILE"
}

test_version() {
    echo -n "Test: --version command... "
    result=$(bash "$CC_SCRIPT" --version 2>&1)
    if echo "$result" | grep -q "cc version"; then
        echo "PASS"
        ((PASS++))
    else
        echo "FAIL (got: $result)"
        ((FAIL++))
    fi
}

test_list() {
    echo -n "Test: --list command... "
    export HOME="."
    export CC_CONFIG_FILE="$CONFIG_FILE"
    result=$(CONFIG_FILE="$CONFIG_FILE" bash -c 'source <(sed "s|\\\$HOME/.cc-config.json|'"$CONFIG_FILE"'|g" '"$CC_SCRIPT"') && list_models' 2>&1)
    if echo "$result" | grep -q "Test Model 1"; then
        echo "PASS"
        ((PASS++))
    else
        echo "FAIL"
        ((FAIL++))
    fi
}

test_help() {
    echo -n "Test: --help command... "
    result=$(bash "$CC_SCRIPT" --help 2>&1)
    if echo "$result" | grep -q "Usage:"; then
        echo "PASS"
        ((PASS++))
    else
        echo "FAIL"
        ((FAIL++))
    fi
}

test_delete_option_exists() {
    echo -n "Test: --delete option in help... "
    result=$(bash "$CC_SCRIPT" --help 2>&1)
    if echo "$result" | grep -q "\-\-delete"; then
        echo "PASS"
        ((PASS++))
    else
        echo "FAIL"
        ((FAIL++))
    fi
}

test_version_option_in_help() {
    echo -n "Test: --version option in help... "
    result=$(bash "$CC_SCRIPT" --help 2>&1)
    if echo "$result" | grep -q "\-\-version"; then
        echo "PASS"
        ((PASS++))
    else
        echo "FAIL"
        ((FAIL++))
    fi
}

echo "======================================="
echo "  CC-CLI Test Suite (Phase 1)"
echo "======================================="
echo ""

setup

test_version
test_help
test_version_option_in_help
test_delete_option_exists

cleanup

echo ""
echo "======================================="
echo "  Results: $PASS passed, $FAIL failed"
echo "======================================="

if [ $FAIL -gt 0 ]; then
    exit 1
fi
