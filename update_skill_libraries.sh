#!/bin/sh
parent="skill-libraries"


for dir in "$parent"/*/; do
  if [ -d "$dir/.git" ]; then
    echo "==> $dir"
    git -C "$dir" reset --hard
    git -C "$dir" clean -fd
    git -C "$dir" pull

    # The pull restores upstream's nested layout, so re-flatten after each update.
    case "$dir" in
      */mattpocock-skills/) sh scripts/flatten_mattpocock_skills.sh "$dir" ;;
    esac
  fi
done
