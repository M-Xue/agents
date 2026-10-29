#!/bin/sh

# Points Claude Code's personal skills directory (~/.claude/skills, or
# $CLAUDE_DIR/skills) at this workspace's skills/ directory, where the skill
# selector links the chosen skills.
#
# Claude Code loads personal skills from ~/.claude/skills only; it does not
# scan ~/.agents/skills. Replacing that directory with a symlink to the
# workspace's skills/ makes every skill the selector exposes available to
# Claude Code too. Anything already at ~/.claude/skills is moved to
# ~/.claude/skills-old rather than deleted; an empty directory is simply
# removed. Re-running is a no-op once the link is in place.
#
# install.sh runs this after building the selector; it can also be run on its
# own from any directory.

set -eu

CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"

# This script sits in skill-selector-app/, whose parent is the agents
# workspace root.
app_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
workspace_root=$(dirname "$app_dir")

exposed_dir="$workspace_root/skills"
claude_skills="$CLAUDE_DIR/skills"
claude_skills_old="$CLAUDE_DIR/skills-old"

# Canonical path of a directory, or empty when it does not resolve to one.
resolve_dir() {
    (CDPATH='' cd -P -- "$1" 2>/dev/null && pwd -P)
}

mkdir -p "$exposed_dir" "$CLAUDE_DIR"

if [ -L "$claude_skills" ] && [ "$(resolve_dir "$claude_skills")" = "$(resolve_dir "$exposed_dir")" ]; then
    printf '%s\n' "correct symlink already exists: $claude_skills -> $exposed_dir"
    exit 0
fi

if [ -L "$claude_skills" ] || [ -e "$claude_skills" ]; then
    if [ ! -L "$claude_skills" ] && [ -d "$claude_skills" ] && [ -z "$(ls -A "$claude_skills")" ]; then
        rmdir "$claude_skills"
    elif [ -L "$claude_skills_old" ] || [ -e "$claude_skills_old" ]; then
        printf '%s\n' "error: $claude_skills is in the way and $claude_skills_old already exists; move one of them aside and re-run" >&2
        exit 1
    else
        mv "$claude_skills" "$claude_skills_old"
        printf '%s\n' "moved existing $claude_skills to $claude_skills_old"
    fi
fi

ln -s "$exposed_dir" "$claude_skills"
printf '%s\n' "linked $claude_skills -> $exposed_dir"
