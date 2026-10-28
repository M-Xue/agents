#!/bin/sh
parent="skill-libraries"

# Reset a library to a clean state and pull the latest changes. On failure
# print a warning and return non-zero so the caller can skip the rest of the
# work for that library while still moving on to the next one.
update() {
  if ! { git -C "$1" reset --hard && git -C "$1" clean -fd && git -C "$1" pull; }; then
    echo "warning: failed to update $1, skipping" >&2
    return 1
  fi
}

for dir in "$parent"/*/; do
  if [ -d "$dir/.git" ]; then
    echo "==> $dir"
    update "$dir" || continue

    # The pull restores upstream's nested layout, so re-flatten after each update.
    case "$dir" in
      */mattpocock-skills/) sh scripts/flatten_mattpocock_skills.sh "$dir" \
        || echo "warning: failed to flatten $dir" >&2 ;;
    esac
  fi
done
