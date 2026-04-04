#!/usr/bin/env bash

set -euo pipefail

REPO_OWNER="LiukerSun"
REPO_NAME="cc-cli"
PROJECT_NAME="ccc"
REPO_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}"

SCRIPT_DIR=""
if [ -n "${BASH_SOURCE[0]-}" ] && [ -f "${BASH_SOURCE[0]}" ]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

VERSION_FILE=""
if [ -n "$SCRIPT_DIR" ]; then
    VERSION_FILE="${SCRIPT_DIR}/VERSION"
fi
LOCAL_VERSION="dev"
if [ -n "$VERSION_FILE" ] && [ -f "${VERSION_FILE}" ]; then
    LOCAL_VERSION="$(tr -d '[:space:]' < "${VERSION_FILE}")"
fi

INSTALL_BIN_DIR="${CCC_INSTALL_BIN_DIR:-${HOME}/.local/bin}"
INSTALL_PATH="${INSTALL_BIN_DIR}/${PROJECT_NAME}"
DATA_DIR="${XDG_DATA_HOME:-${HOME}/.local/share}/${PROJECT_NAME}"
RELEASES_DIR="${DATA_DIR}/releases"

ACTION="install"
REQUESTED_VERSION="${CCC_VERSION:-latest}"
NO_SHELL_CONFIG="${CCC_NO_SHELL_CONFIG:-}"

usage() {
    cat <<EOF
Install ${PROJECT_NAME} from GitHub Releases.

Usage:
  ./install.sh
  ./install.sh --version 2.2.1
  ./install.sh --uninstall
  ./install.sh --no-shell-config

Environment:
  CCC_VERSION            Version to install. Defaults to latest.
  CCC_INSTALL_BIN_DIR    Override the install directory. Defaults to ~/.local/bin.
  CCC_NO_SHELL_CONFIG    Set to any value to skip shell configuration.

Notes:
  - The installer installs the ${PROJECT_NAME} binary only.
  - Node.js / npm are no longer required to install ${PROJECT_NAME} itself.
  - Missing target CLIs (claude / codex) are handled at runtime by ${PROJECT_NAME}.
EOF
}

log() {
    printf '%s\n' "$*"
}

fail() {
    printf 'Error: %s\n' "$*" >&2
    exit 1
}

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --help|-h)
                usage
                exit 0
                ;;
            --uninstall)
                ACTION="uninstall"
                ;;
            --no-shell-config)
                NO_SHELL_CONFIG=1
                ;;
            --version)
                [ $# -ge 2 ] || fail "--version requires a value"
                REQUESTED_VERSION="$2"
                shift
                ;;
            *)
                fail "unknown argument: $1"
                ;;
        esac
        shift
    done
}

detect_os() {
    case "$(uname -s)" in
        Linux) echo "linux" ;;
        Darwin) echo "darwin" ;;
        *)
            fail "unsupported operating system: $(uname -s)"
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)
            fail "unsupported architecture: $(uname -m)"
            ;;
    esac
}

# Detect shell type from $SHELL environment variable
detect_shell() {
    local shell_path="${SHELL:-}"
    if [ -z "$shell_path" ]; then
        echo "bash"
        return
    fi
    local shell_name
    shell_name="$(basename "$shell_path")"
    case "$shell_name" in
        bash|zsh|fish)
            echo "$shell_name"
            ;;
        *)
            echo "bash"
            ;;
    esac
}

# Get shell config file path based on shell type and OS
get_shell_config_file() {
    local shell_type="$1"
    local os
    os="$(detect_os)"

    case "$shell_type" in
        bash)
            if [ "$os" = "darwin" ]; then
                echo "${HOME}/.bash_profile"
            else
                echo "${HOME}/.bashrc"
            fi
            ;;
        zsh)
            echo "${HOME}/.zshrc"
            ;;
        fish)
            echo "${HOME}/.config/fish/config.fish"
            ;;
        *)
            echo "${HOME}/.bashrc"
            ;;
    esac
}

# Check if PATH already exists in config file
is_path_in_config() {
    local config_file="$1"
    local path_to_check="$2"

    if [ ! -f "$config_file" ]; then
        return 1
    fi

    if grep -qF "$path_to_check" "$config_file" 2>/dev/null; then
        return 0
    fi

    return 1
}

# Add PATH to bash/zsh config
add_path_to_bash_config() {
    local config_file="$1"
    local bin_dir="$2"

    # Create config file if it doesn't exist
    local config_dir
    config_dir="$(dirname "$config_file")"
    mkdir -p "$config_dir"

    # Append PATH export to config file
    {
        echo ""
        echo "# Added by ${PROJECT_NAME} installer"
        echo "export PATH=\"${bin_dir}:\$PATH\""
    } >> "$config_file"
}

# Add PATH to fish config
add_path_to_fish_config() {
    local config_file="$1"
    local bin_dir="$2"

    # Create config file if it doesn't exist
    local config_dir
    config_dir="$(dirname "$config_file")"
    mkdir -p "$config_dir"

    # Append fish_add_path to config file
    {
        echo ""
        echo "# Added by ${PROJECT_NAME} installer"
        echo "fish_add_path ${bin_dir}"
    } >> "$config_file"
}

# Configure shell by updating config file
configure_shell() {
    # Skip if NO_SHELL_CONFIG is set
    if [ -n "$NO_SHELL_CONFIG" ]; then
        return 0
    fi

    # Skip if install directory is already in PATH
    case ":${PATH}:" in
        *":${INSTALL_BIN_DIR}:"*)
            log ""
            log "${INSTALL_BIN_DIR} is already in PATH."
            return 0
            ;;
    esac

    local shell_type
    shell_type="$(detect_shell)"
    local config_file
    config_file="$(get_shell_config_file "$shell_type")"

    # Check if already configured
    if is_path_in_config "$config_file" "$INSTALL_BIN_DIR"; then
        log ""
        log "${INSTALL_BIN_DIR} is already configured in ${config_file}."
        log "Please reload your shell configuration or open a new terminal."
        return 0
    fi

    # Add PATH to config file
    case "$shell_type" in
        fish)
            add_path_to_fish_config "$config_file" "$INSTALL_BIN_DIR"
            ;;
        *)
            add_path_to_bash_config "$config_file" "$INSTALL_BIN_DIR"
            ;;
    esac

    log ""
    log "Updated ${config_file} to add ${INSTALL_BIN_DIR} to PATH."
    log "Please reload your shell configuration:"
    case "$shell_type" in
        bash)
            log "  source ${config_file}"
            ;;
        zsh)
            log "  source ${config_file}"
            ;;
        fish)
            log "  source ${config_file}"
            ;;
    esac
    log "Or open a new terminal."
}

normalize_tag() {
    if [ "$REQUESTED_VERSION" = "latest" ]; then
        echo "latest"
    elif [[ "$REQUESTED_VERSION" == v* ]]; then
        echo "$REQUESTED_VERSION"
    else
        echo "v${REQUESTED_VERSION}"
    fi
}

download() {
    local url="$1"
    local output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$output" "$url"
    else
        fail "curl or wget is required to download release assets"
    fi
}

verify_checksum() {
    local asset_path="$1"
    local checksum_path="$2"
    local asset_name
    asset_name="$(basename "$asset_path")"

    if [ ! -f "$checksum_path" ]; then
        return 0
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$(dirname "$asset_path")" && grep " ${asset_name}\$" "$checksum_path" | sha256sum -c - >/dev/null)
        return 0
    fi

    if command -v shasum >/dev/null 2>&1; then
        local expected
        expected="$(grep " ${asset_name}\$" "$checksum_path" | awk '{print $1}')"
        [ -n "$expected" ] || return 0
        local actual
        actual="$(shasum -a 256 "$asset_path" | awk '{print $1}')"
        [ "$expected" = "$actual" ] || fail "checksum verification failed for ${asset_name}"
    fi
}

install_binary() {
    local built_binary="$1"
    mkdir -p "$INSTALL_BIN_DIR" "$RELEASES_DIR"
    if command -v install >/dev/null 2>&1; then
        install -m 755 "$built_binary" "$INSTALL_PATH"
    else
        cp "$built_binary" "$INSTALL_PATH"
        chmod 755 "$INSTALL_PATH"
    fi
}

print_path_hint() {
    case ":$PATH:" in
        *":${INSTALL_BIN_DIR}:"*)
            ;;
        *)
            log ""
            log "Warning: ${INSTALL_BIN_DIR} is not in PATH."
            log "Add this to your shell profile:"
            log "  export PATH=\"${INSTALL_BIN_DIR}:\$PATH\""
            ;;
    esac
}

install_from_local_checkout() {
    command -v go >/dev/null 2>&1 || fail "go is required to build from a local checkout"

    local tmp_dir
    tmp_dir="$(mktemp -d)"

    log "Building ${PROJECT_NAME} from local checkout..."
    (
        cd "$SCRIPT_DIR"
        go build -trimpath -ldflags "-X github.com/LiukerSun/cc-cli/internal/buildinfo.Version=${LOCAL_VERSION}" -o "${tmp_dir}/${PROJECT_NAME}" ./cmd/ccc
    )

    install_binary "${tmp_dir}/${PROJECT_NAME}"
    rm -rf "${tmp_dir}"
}

install_from_release() {
    local os arch tag asset_name asset_url checksum_url tmp_dir archive_path checksum_path extracted_path base_url
    os="$(detect_os)"
    arch="$(detect_arch)"
    tag="$(normalize_tag)"
    asset_name="${PROJECT_NAME}_${os}_${arch}.tar.gz"

    if [ "$tag" = "latest" ]; then
        base_url="${REPO_URL}/releases/latest/download"
    else
        base_url="${REPO_URL}/releases/download/${tag}"
    fi

    asset_url="${base_url}/${asset_name}"
    checksum_url="${base_url}/checksums.txt"
    tmp_dir="$(mktemp -d)"

    archive_path="${tmp_dir}/${asset_name}"
    checksum_path="${tmp_dir}/checksums.txt"

    log "Downloading ${asset_name}..."
    download "$asset_url" "$archive_path"
    download "$checksum_url" "$checksum_path" || true
    verify_checksum "$archive_path" "$checksum_path"

    tar -xzf "$archive_path" -C "$tmp_dir"
    extracted_path="${tmp_dir}/${PROJECT_NAME}"
    [ -f "$extracted_path" ] || fail "release archive did not contain ${PROJECT_NAME}"

    install_binary "$extracted_path"
    rm -rf "${tmp_dir}"
}

uninstall() {
    local removed=0
    for path in \
        "$INSTALL_PATH" \
        "${HOME}/bin/${PROJECT_NAME}" \
        "${HOME}/.ccc/bin/${PROJECT_NAME}"
    do
        if [ -e "$path" ] || [ -L "$path" ]; then
            rm -f "$path"
            log "Removed ${path}"
            removed=1
        fi
    done

    if [ "$removed" -eq 0 ]; then
        log "${PROJECT_NAME} is not installed in the known locations."
    fi

    log "Config and data were left untouched."
}

main() {
    parse_args "$@"

    if [ "$ACTION" = "uninstall" ]; then
        uninstall
        exit 0
    fi

    if [ -n "$SCRIPT_DIR" ] && [ -f "${SCRIPT_DIR}/go.mod" ] && [ -d "${SCRIPT_DIR}/cmd/ccc" ] && { [ "$REQUESTED_VERSION" = "latest" ] || [ "$REQUESTED_VERSION" = "$LOCAL_VERSION" ] || [ "$REQUESTED_VERSION" = "v${LOCAL_VERSION}" ]; }; then
        install_from_local_checkout
    else
        install_from_release
    fi

    log ""
    log "${PROJECT_NAME} installed to ${INSTALL_PATH}"
    configure_shell
    print_path_hint
    log ""
    log "Run '${PROJECT_NAME} version' to verify the installation."
}

main "$@"
