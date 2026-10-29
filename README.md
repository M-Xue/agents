# agents

> **Clone this repo to `~/.agents`.** The skill selector links selected
> skills into `skills/` at the root of this repo, so the checkout must live at
> `~/.agents` for that directory to be `~/.agents/skills`, the path most
> coding agents read skills from. Cloned anywhere else, the links are created
> but those agents do not see them. Claude Code is the exception: it reads
> `~/.claude/skills` only, so the selector's installer points that directory
> at this repo's `skills/` (see [Installing the skill
> selector](#installing-the-skill-selector)).
>
> ```bash
> git clone <repository url> ~/.agents
> ```

A workspace for managing the skills exposed to coding agents. Skills are
collected from skill libraries and custom sources, then selectively
linked into place with a terminal UI.

## Custom Skills

Foo

## Skill Libraries

Skill libraries are cloned by `install_skill_libraries.sh` into
the `skill-libraries/` directory. That directory is gitignored, so it acts
as a local library to pull skills from rather than something committed to this
repo. Run `update_skill_libraries.sh` to pull the latest changes for every
cloned library.

Both scripts treat each library independently: if a clone or update fails
(for example because the directory already exists from an earlier run, or the
network is unavailable), the script prints a warning to stderr and carries on
with the remaining libraries rather than aborting.

### Adding new skill libraries

A skill library is a git repository with a `skills/` folder at the root of the
directory. To add a library, add a `clone` line to
`install_skill_libraries.sh` in the following structure:

```sh
clone <repository url> skill-libraries/<name of library>
```

`clone` is a small helper defined at the top of the script that wraps
`git clone` and turns a failure into a warning instead of a hard stop.

### Expected layout

The skill selector expects every library to keep its skills in a flat
`skills/` directory at the library root, one subdirectory per skill, each
containing a `SKILL.md`:

```txt
<library>/
└── skills/
    ├── skill-a/
    │   └── SKILL.md
    └── skill-b/
        └── SKILL.md
```

Matt Pocock's library does not follow this structure.
It groups skills under category directories such as
`skills/engineering/` and `skills/productivity/`. To bring it in line, we run
`scripts/flatten_mattpocock_skills.sh` in the install and update script.
The script moves each skill up to the root of `skills/`, removes the empty
category directories, and refuses to run if two skills would end up with the
same name.

### Selecting skills

Once the libraries are installed, use the skill selector, a Go TUI in
`skill-selector-app/`, to choose which skills are exposed to the agent at any
given time. Selected skills are exposed via symlinks into `skills/`, so
toggling a skill on or off only adds or removes a link rather than copying
files. See [Installing the skill selector](#installing-the-skill-selector)
to set it up, then run `skills` from anywhere.

The Skill Libraries panel lists every cloned library plus two built-in
entries at the top. `All`, selected on startup, combines the
skills of every library into one list. `Currently selected`,
just beneath it, lists only the skills that are exposed right now, so it is
the quickest way to review or prune your selection: turning a skill off
there removes it from the list immediately, and toggle all turns everything
off. In both views each skill is tagged with its library name, since
different libraries can ship skills with the same name, and searching matches
library names too, so typing a library's name narrows the list to its skills.

Because a skill is exposed under its directory name, only one skill of a
given name can be on at a time. When one is, the same-named skills in other
libraries are shown dimmed with a `[skill with same name already exists]`
suffix; toggling one of them explains which library
already holds the name, and toggle all (ctrl+a) treats them as already
toggled, so it still switches everything else off and on cleanly.

## Installing the skill selector

Requirements:

- Go 1.27+
- A [Nerd Font](https://www.nerdfonts.com/) compatible terminal font. The
  checkbox icons in the Skills panel are Nerd Font glyphs and render as
  blank or placeholder boxes without one.

The installer builds the binary and places it on your path. It locates this
repo itself, so you can run it from any directory:

```bash
sh skill-selector-app/install.sh
```

This builds the selector and installs it as `skills` at `~/.local/bin/skills`
(honoring `$BIN_DIR`). The path of this workspace is compiled into the binary,
so `skills` finds the right `skills/` and `skill-libraries/`
directories no matter where you run it from. If `~/.local/bin` is not already
on your `PATH`, the script tells you the line to add to your shell startup
file.

The installer also wires up Claude Code by running
`skill-selector-app/link_claude_skills.sh`. Claude Code loads personal skills
from `~/.claude/skills` and does not scan `~/.agents/skills`, so that script
replaces `~/.claude/skills` with a symlink to this repo's `skills/`
directory. Every skill you expose with the selector is then available in
Claude Code too. If `~/.claude/skills` already holds skills, they are moved
to `~/.claude/skills-old` rather than deleted; link the ones you still want
back into `skills/`, or re-link them from a library, and delete the old
directory when you are done. An empty `~/.claude/skills` is simply removed.
Set `$CLAUDE_DIR` to use a Claude config directory other than `~/.claude`.
The step is skipped when the link is already in place, and it refuses to run
if both a real `~/.claude/skills` and a `~/.claude/skills-old` exist, so a
previous backup is never overwritten. The script can also be run on its own,
without rebuilding the selector:

```bash
sh skill-selector-app/link_claude_skills.sh
```

When started inside a different checkout of this repo, the selector prefers
that checkout over the compiled-in one, and `skills -root <path>`
overrides both.

### Updating

Re-run the installer after pulling a new version. It rebuilds into the same
location and is idempotent.

### Uninstall

```bash
sh skill-selector-app/uninstall.sh
```

This removes the binary `install.sh` created. It honors the same `$BIN_DIR`
override, so export it again here if you set it at install time. The symlinks
in `skills/` are your selection, not part of the program, and are left in
place; remove them with the selector itself or by hand. The
`~/.claude/skills` link is left in place too, so Claude Code keeps seeing the
skills in `skills/`; remove it and restore `~/.claude/skills-old` by hand if
you want the previous layout back.
