#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION_FILE="$REPO_DIR/VERSION"

if [ ! -f "$VERSION_FILE" ]; then
    echo "VERSION file not found: $VERSION_FILE" >&2
    exit 1
fi

VERSION="$(tr -d '[:space:]' < "$VERSION_FILE")"

if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "VERSION must use semantic version format MAJOR.MINOR.PATCH, got: $VERSION" >&2
    exit 1
fi

TAG_INPUT="${1:-${GITHUB_REF_NAME:-}}"
if [ -n "$TAG_INPUT" ]; then
    TAG="${TAG_INPUT#refs/tags/}"
    TAG="${TAG#v}"
    if [ "$VERSION" != "$TAG" ]; then
        echo "VERSION ($VERSION) does not match release tag (${TAG_INPUT})" >&2
        exit 1
    fi
fi

echo "check_version.sh: ok ($VERSION)"
