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
    │   ├── model.go            root Model, Init/Update/View
    │   ├── layout.go           window frame: padding, footer rows, side-by-side panels
    │   ├── keys.go             key bindings rendered in the help footer
    │   ├── theme/
    │   │   └── theme.go        shared chrome colours (borders, title tab)
    │   └── component/          independent presentation components, one per folder
    │       ├── doc.go          the component rules
    │       ├── panel/
    │       │   └── panel.go    titled sharp-cornered box: New(title, w, h), View(content)
    │       └── checklist/
    │           └── checklist.go scrollable checked/clear list with cursor
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

The visual chrome is a port of the grove TUI (`~/Experiments/grove`) and is
split into two subpackages so screens stay free of styling detail:

- `ui/theme` holds the colour palette. The active border and title tab are
  ANSI 111 (light blue), inactive borders are ANSI 60 (slate), tab text is
  ANSI 235 (near black).
- `ui/component/panel` draws a lazygit-style pane: square corners, the title
  rendered as a bold highlighted tab embedded in the top border (`┌─ Title ──┐`),
  content padded one column from each side, always filled to the exact
  width and height requested. `panel.Tab` renders the standalone app title in
  the same style.
- `ui/component/checklist` is a scrolling list of `Item{ID, Checked, Label}`
  rows. Unchecked rows show U+E640, checked rows U+F4A7 (both Nerd Font
  glyphs), the cursor row is bold. The component owns the cursor and scroll
  offset and wraps at both ends; checked state is passed in by the caller
  and never mutated by the component.

#### Component rules

Every folder under `ui/component` is one independent component. The rules
are strict so components stay reusable and trivially testable:

1. A component imports nothing from this project except `ui/theme`.
2. A component never imports another component. If two need to share code,
   that code is duplicated or lifted into the caller, not shared sideways.
3. A component contains no business logic. It knows nothing about skills,
   files, symlinks or the app's state.
4. A component receives everything it needs as arguments. Width, height,
   title, focus and content are computed by the caller and passed in. A
   component never reads the terminal size or decides its own layout.
5. A component follows the Bubble Tea shape where it has state: `New(...)`
   constructor, setters, and a `View` that returns a string. Components with
   no state are plain functions.

The caller, usually a screen or the root model, owns layout: it decides how
many columns each component gets and tells it.

`layout.go` owns the window frame. The content area is inset 4 columns on the
left and right and 2 rows on the top and bottom. Inside it, from top to
bottom: the app tab, a blank line, the panels, then a single reserved row for
the help footer. The panels run right down to the footer with no gap. The three panels, Skill Libraries, Skills and Details, are
split 1:1:2 with a single blank column between each. Tab and shift+tab move
focus between them; the focused panel takes the active border colour. The footer is rendered with
`bubbles/help` in short mode from the bindings in `keys.go`.

### `internal/app`

General business logic. It coordinates the `skills` and `files` packages and
exposes the operations the UI needs. It knows nothing about the terminal and
can be driven from tests or a future CLI without changes.

Selection state lives here, not in the UI. The checklist component reports
which row is under the cursor; `app.ToggleSkill` flips the selection and the
UI re-reads `app.Skills` plus `app.IsSelected` to rebuild the rows. The
current skill list is hard-coded placeholder data to be replaced once the
`skills` package discovers real directories.

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
- Third-party dependencies are added when first used. The `ui` package pulls
  in `charmbracelet/bubbletea`, `charmbracelet/lipgloss` and
  `charmbracelet/bubbles`; nothing outside `ui` may import them.
