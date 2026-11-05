#!/bin/sh

# Points Claude Code's personal config at this workspace:
#
#   $CLAUDE_DIR/skills     -> <workspace>/skills     (default CLAUDE_DIR=~/.claude)
#   $CLAUDE_DIR/CLAUDE.md  -> <workspace>/AGENTS.md
#
# Claude Code loads personal skills from ~/.claude/skills only; it does not
# scan ~/.agents/skills. Replacing that directory with a symlink to the
# workspace's skills/ makes every skill the selector exposes available to
# Claude Code too. Anything already at ~/.claude/skills is moved to
# ~/.claude/skills-old rather than deleted; an empty directory is simply
# removed.
#
# Likewise Claude Code reads global instructions from ~/.claude/CLAUDE.md,
# while other agents read AGENTS.md. Linking the former to the workspace's
# AGENTS.md keeps a single copy. If AGENTS.md does not exist yet, an existing
# CLAUDE.md is moved there so its content is kept; failing that an empty
# AGENTS.md is created. If both exist, CLAUDE.md is moved to CLAUDE.md-old
# rather than deleted.
#
# Re-running is a no-op once the links are in place.
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

agents_md="$workspace_root/AGENTS.md"
claude_md="$CLAUDE_DIR/CLAUDE.md"
claude_md_old="$CLAUDE_DIR/CLAUDE.md-old"

# Canonical path of a directory, or empty when it does not resolve to one.
resolve_dir() {
    (CDPATH='' cd -P -- "$1" 2>/dev/null && pwd -P)
}


mkdir -p "$exposed_dir" "$CLAUDE_DIR"

# --- skills directory -------------------------------------------------------

if [ -L "$claude_skills" ] && [ "$(resolve_dir "$claude_skills")" = "$(resolve_dir "$exposed_dir")" ]; then
    printf '%s\n' "correct symlink already exists: $claude_skills -> $exposed_dir"
else
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
fi

# --- CLAUDE.md -> AGENTS.md -------------------------------------------------

# Already linked to the right place: nothing to do. Compare the raw link
# target, since AGENTS.md is a file and may not exist yet.
if [ -L "$claude_md" ] && [ "$(readlink "$claude_md")" = "$agents_md" ]; then
    printf '%s\n' "correct symlink already exists: $claude_md -> $agents_md"
    exit 0
fi

if [ -L "$claude_md" ] || [ -e "$claude_md" ]; then
    if [ -L "$claude_md" ]; then
        # A link to somewhere else; drop it, nothing of value is lost.
        rm "$claude_md"
        printf '%s\n' "removed stale symlink $claude_md"
    elif [ ! -f "$claude_md" ]; then
        printf '%s\n' "error: $claude_md exists but is not a regular file; move it aside and re-run" >&2
        exit 1
    elif [ ! -e "$agents_md" ]; then
        # No AGENTS.md yet: adopt the existing CLAUDE.md as it.
        mv "$claude_md" "$agents_md"
        printf '%s\n' "moved existing $claude_md to $agents_md"
    elif [ -L "$claude_md_old" ] || [ -e "$claude_md_old" ]; then
        printf '%s\n' "error: $claude_md is in the way and $claude_md_old already exists; move one of them aside and re-run" >&2
        exit 1
    else
        mv "$claude_md" "$claude_md_old"
        printf '%s\n' "moved existing $claude_md to $claude_md_old"
    fi
fi

if [ ! -e "$agents_md" ]; then
    : > "$agents_md"
    printf '%s\n' "created empty $agents_md"
fi

ln -s "$agents_md" "$claude_md"
printf '%s\n' "linked $claude_md -> $agents_md"
