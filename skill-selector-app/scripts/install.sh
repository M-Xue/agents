#!/bin/sh

# Builds skill-selector and installs it to $BIN_DIR (default ~/.local/bin).
# The agents workspace this checkout lives in is compiled into the binary, so
# the installed command works from any directory. It is safe to re-run: each
# run rebuilds into the same location.

set -eu

BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
BINARY_PATH="$BIN_DIR/skill-selector"

PLATFORM="$(uname -s)"
case "$PLATFORM" in
    Darwin|Linux)
        ;;
    *)
        printf '%s\n' "unsupported platform: $PLATFORM" >&2
        exit 1
        ;;
esac

if ! command -v go >/dev/null 2>&1; then
    printf '%s\n' "go is required to install skill-selector" >&2
    exit 1
fi

# Resolve paths from the script's own location so this works from anywhere:
# scripts/ -> skill-selector-app/ -> the agents workspace root.
script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
app_dir=$(dirname "$script_dir")
workspace_root=$(dirname "$app_dir")

mkdir -p "$BIN_DIR"
(cd "$app_dir" && go build -ldflags "-X main.defaultRoot=$workspace_root" -o "$BINARY_PATH" .)
chmod +x "$BINARY_PATH"

printf '%s\n' "installed skill-selector to $BINARY_PATH (workspace: $workspace_root)"

case ":${PATH:-}:" in
    *":$BIN_DIR:"*)
        ;;
    *)
        printf '%s\n' "note: $BIN_DIR is not on your PATH; add it to your shell startup file, e.g.:" >&2
        printf '%s\n' "  export PATH=\"$BIN_DIR:\$PATH\"" >&2
        ;;
esac
