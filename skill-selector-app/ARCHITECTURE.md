# Architecture

skill-selector is a terminal UI for discovering agent skill directories and
linking them into place. The code is split into four packages with a strict
one-way dependency flow.

## Layout

```
skill-selector-app/
├── go.mod                      module github.com/maxx/skill-selector-app
├── .gitignore
├── cmd/skill-selector/
│   └── main.go                 entry point, wires packages together
└── internal/
    ├── ui/                     Bubble Tea TUI (model, update, view)
    │   ├── doc.go
    │   └── model.go            Model struct, New()
    ├── app/                    general business logic, coordinates skills + files
    │   ├── doc.go
    │   └── app.go              App struct, New()
    ├── skills/                 processes agent skill directories
    │   ├── doc.go
    │   └── skill.go            Skill struct
    └── files/                  filesystem access only, no business logic
        ├── doc.go
        ├── read.go             Read(path)
        ├── symlink.go          CreateSymlink, IsSymlink, IsSymlinkToDir
        └── list.go             Entry struct, List(root)
```

## Packages

### `cmd/skill-selector`

The `main` package. It constructs the `app.App`, builds the `ui.Model` and
starts the Bubble Tea program. It contains no logic of its own.

### `internal/ui`

Owns everything Bubble Tea: the root model, `Init`, `Update`, `View`, key
handling and rendering. It never makes decisions about skills or files; it
calls into `app` and displays the result.

### `internal/app`

General business logic. It coordinates the `skills` and `files` packages and
exposes the operations the UI needs. It knows nothing about the terminal and
can be driven from tests or a future CLI without changes.

### `internal/skills`

Processes agent skill directories. A skill is a directory containing a
`SKILL.md` and supporting files. This package discovers, parses and describes
skills. It does not touch the filesystem directly; it relies on `files`.

### `internal/files`

Thin wrapper over filesystem access:

- reading files
- creating symlinks
- checking whether a path is a symlink, and whether it points at a directory
- listing all files and directories beneath a given subfolder

No business logic lives here. Symlink checks use `os.Lstat` so that a link to
a directory is reported as a link, not as a directory.

## Dependency direction

```
cmd/skill-selector
       │
       ▼
      ui ──▶ app ──▶ skills ──▶ files
              │                  ▲
              └──────────────────┘
```

Dependencies only point downward. `files` imports nothing from the project,
`skills` may import `files`, `app` may import `skills` and `files`, and `ui`
imports only `app`. Nothing imports `ui`.

## Conventions

- Packages live under `internal/` so they cannot be imported by other modules.
- Unimplemented functions panic with `not implemented` rather than returning
  zero values, so an accidental call fails loudly.
- Each package has a `doc.go` stating its responsibility and what it must not
  do.
- Third-party dependencies are added when first used. Bubble Tea is added with
  `go get github.com/charmbracelet/bubbletea` when work on `ui` begins.
