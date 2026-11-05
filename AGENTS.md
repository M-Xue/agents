# Agent rules

## Git
- Never run `git commit` or `git push`. The only exception is the `submit-pr` skill and the skills it invokes (`organise-changes-into-stacked-prs`, `write-pr-description`), which own committing and pushing. Otherwise leave changes in the working tree and say they are ready.
- Never edit a PR title or description unless asked explicitly. The `submit-pr` skill writing the description of a PR it is opening is the one exception.

## Code style
- Prefer self-explanatory code over comments. Rename, extract or restructure first. Add a comment only for something the code cannot express, such as a workaround or a non-obvious constraint.
