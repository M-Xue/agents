# agents

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

Each clone is independent: if one fails (for example because the directory
already exists from an earlier run, or the network is unavailable), the script
prints a warning to stderr and carries on with the remaining libraries rather
than aborting.

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
place; remove them with the selector itself or by hand.
