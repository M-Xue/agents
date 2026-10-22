# agents

A workspace for managing the skills exposed to coding agents. Skills are
collected from third-party libraries and custom sources, then selectively
linked into place with a terminal UI.

## Custom Skills

Foo

## Third Party Skills

Third-party skill libraries are cloned by `install_third_party_skills.sh` into
the `third-party-skills/` directory. That directory is gitignored, so it acts
as a local library to pull skills from rather than something committed to this
repo. Run `update_third_party_skills.sh` to pull the latest changes for every
cloned library.

Currently installed libraries:

- [pstack](https://github.com/backnotprop/pstack) (backnotprop)
- [skills](https://github.com/mattpocock/skills) (mattpocock)
- [agent-skills](https://github.com/addyosmani/agent-skills) (addyosmani)
- [ponytail](https://github.com/DietrichGebert/ponytail) (DietrichGebert)
- [superpowers](https://github.com/obra/superpowers) (obra)

### Expected layout

The skill selector expects every library to keep its skills in a flat
`skills/` directory at the library root, one subdirectory per skill, each
containing a `SKILL.md`:

```
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
given time. Selected skills are exposed via symlinks, so toggling a skill on or
off only adds or removes a link rather than copying files.
