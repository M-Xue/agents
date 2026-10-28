#!/bin/sh

# Clone a repo into a destination directory. On failure (e.g. the directory
# already exists from an earlier run, or the network is down) print a warning
# and keep going so the remaining libraries are still installed.
clone() {
  if ! git clone "$1" "$2"; then
    echo "warning: failed to clone $1 into $2, skipping" >&2
    return 1
  fi
}

clone https://github.com/backnotprop/pstack.git skill-libraries/pstack
# Matt Pocock's library nests skills under category directories, so flatten it
# once the clone succeeds. A failed clone leaves the existing checkout untouched.
if clone https://github.com/mattpocock/skills.git skill-libraries/mattpocock-skills; then
  sh scripts/flatten_mattpocock_skills.sh skill-libraries/mattpocock-skills \
    || echo "warning: failed to flatten skill-libraries/mattpocock-skills" >&2
fi
clone https://github.com/addyosmani/agent-skills.git skill-libraries/addy-skills
clone https://github.com/DietrichGebert/ponytail.git skill-libraries/ponytail
clone https://github.com/obra/superpowers.git skill-libraries/superpowers
clone https://github.com/M-Xue/skills.git skill-libraries/max-skills
clone https://github.com/canvanauts/maxx-canva-skills.git skill-libraries/max-canva-skills
