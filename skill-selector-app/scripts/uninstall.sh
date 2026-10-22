#!/bin/sh

# Removes the binary install.sh created. Honors the same BIN_DIR override, so
# export it here if you set it at install time. The symlinks the selector made
# in the workspace's skills/ directory are left in place: they are your
# selection, not part of the program.

set -eu

BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
BINARY_PATH="$BIN_DIR/skills"

if [ -e "$BINARY_PATH" ]; then
    rm -f "$BINARY_PATH"
    printf '%s\n' "removed $BINARY_PATH"
else
    printf '%s\n' "nothing to remove: $BINARY_PATH does not exist"
fi

printf '%s\n' "skills uninstalled"
