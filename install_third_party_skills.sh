#!/bin/sh

git clone https://github.com/backnotprop/pstack.git third-party-skills/pstack
# Matt Pocock's library nests skills under category directories, so flatten it
# once the clone succeeds. A failed clone (e.g. the directory already exists
# from an earlier run) leaves the existing checkout untouched.
if git clone https://github.com/mattpocock/skills.git third-party-skills/mattpocock-skills; then
  sh scripts/flatten_mattpocock_skills.sh third-party-skills/mattpocock-skills
fi
git clone https://github.com/addyosmani/agent-skills.git third-party-skills/addy-skills
git clone https://github.com/DietrichGebert/ponytail.git third-party-skills/ponytail
git clone https://github.com/obra/superpowers.git third-party-skills/superpowers

