#!/bin/sh

git clone https://github.com/backnotprop/pstack.git skill-libraries/pstack
# Matt Pocock's library nests skills under category directories, so flatten it
# once the clone succeeds. A failed clone (e.g. the directory already exists
# from an earlier run) leaves the existing checkout untouched.
if git clone https://github.com/mattpocock/skills.git skill-libraries/mattpocock-skills; then
  sh scripts/flatten_mattpocock_skills.sh skill-libraries/mattpocock-skills
fi
git clone https://github.com/addyosmani/agent-skills.git skill-libraries/addy-skills
git clone https://github.com/DietrichGebert/ponytail.git skill-libraries/ponytail
git clone https://github.com/obra/superpowers.git skill-libraries/superpowers

