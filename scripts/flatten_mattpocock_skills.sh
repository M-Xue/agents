#!/bin/sh
# Flattens the mattpocock skills library into the layout the skill selector
# expects: every skill directory (one containing a SKILL.md) sits directly
# under skills/ at the library root.
#
# Upstream nests skills under category directories such as
# skills/engineering/<skill> and skills/productivity/<skill>. This script moves
# each skill up one level and removes the now-empty category directories.
#
# It is safe to re-run: skills already at the root are left alone, and the
# script aborts before moving anything if a flattened name would collide.
set -eu

lib="${1:-skill-libraries/mattpocock-skills}"
lib="${lib%/}"
skills="$lib/skills"

if [ ! -d "$skills" ]; then
  echo "error: $skills does not exist. Run install_skill_libraries.sh first." >&2
  exit 1
fi

# Pass 1: check for name collisions before touching anything.
for category in "$skills"/*/; do
  [ -d "$category" ] || continue
  for skill in "$category"*/; do
    [ -f "$skill/SKILL.md" ] || continue
    name="$(basename "$skill")"
    if [ -e "$skills/$name" ]; then
      echo "error: $skills/$name already exists, refusing to overwrite with $skill" >&2
      exit 1
    fi
  done
done

# Pass 2: move each skill up to the root of skills/.
for category in "$skills"/*/; do
  [ -d "$category" ] || continue
  # A category dir that is itself a skill is already flat; leave it.
  [ -f "$category/SKILL.md" ] && continue

  moved=0
  for skill in "$category"*/; do
    [ -f "$skill/SKILL.md" ] || continue
    name="$(basename "$skill")"
    mv "$skill" "$skills/$name"
    echo "moved $(basename "$category")/$name -> $name"
    moved=1
  done

  # Drop the category directory and its README now that it holds no skills.
  # Categories with no skills at all (e.g. deprecated/) are removed too.
  remaining="$(find "$category" -mindepth 1 -type d | head -n 1)"
  if [ -z "$remaining" ]; then
    rm -rf "$category"
    echo "removed $(basename "$category")/"
  elif [ "$moved" -eq 1 ]; then
    echo "warning: $(basename "$category")/ still contains directories, left in place" >&2
  fi
done

echo "done: $skills is flat"
