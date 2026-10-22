#!/bin/sh
parent="third-party-skills"


for dir in "$parent"/*/; do
  if [ -d "$dir/.git" ]; then
    echo "==> $dir"
    git -C "$dir" pull
  fi
done
