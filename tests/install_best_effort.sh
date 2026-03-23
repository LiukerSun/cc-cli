#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/bin" "$TMP_DIR/home/.npm-global/bin"

cat > "$TMP_DIR/bin/node" <<'EOF'
#!/bin/bash
echo v16.20.0
EOF

cat > "$TMP_DIR/bin/npm" <<'EOF'
#!/bin/bash
set -e

if [ "$1" = "config" ] && [ "$2" = "get" ] && [ "$3" = "prefix" ]; then
    echo "$HOME/.npm-global"
    exit 0
fi

if [ "$1" = "--version" ]; then
    echo 10.8.0
    exit 0
fi

if [ "$1" = "install" ] && [ "$2" = "-g" ] && [ "$3" = "@openai/codex" ]; then
    cat > "$HOME/.npm-global/bin/codex" <<'INNER'
#!/bin/bash
exit 0
INNER
    chmod +x "$HOME/.npm-global/bin/codex"
    exit 0
fi

echo "unexpected npm args: $*" >&2
exit 1
EOF

chmod +x "$TMP_DIR/bin/node" "$TMP_DIR/bin/npm"

HOME="$TMP_DIR/home" PATH="$TMP_DIR/bin:/usr/bin:/bin" bash "$REPO_DIR/install.sh" > "$TMP_DIR/output.txt" 2>&1

if ! grep -q "Skipping automatic install for claude: Node.js version is too old" "$TMP_DIR/output.txt"; then
    echo "expected best-effort skip message for claude" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if ! grep -q "codex CLI installed" "$TMP_DIR/output.txt"; then
    echo "expected codex installation attempt to succeed" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if [ ! -x "$TMP_DIR/home/bin/ccc" ]; then
    echo "expected ccc to be installed even when claude auto-install is skipped" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

echo "install_best_effort.sh: ok"
