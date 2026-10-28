# Architecture

skill-selector is a terminal UI for discovering agent skill directories and
linking them into place. The code is split into four packages with a strict
one-way dependency flow.

## Building and installing

The `main` package lives at the module root, so a plain `go build` in this
directory produces `./skill-selector-app`. For day-to-day use run the
installer instead, which builds the binary to `~/.local/bin/skills`
(honouring `$BIN_DIR`) and compiles in the path of the workspace this checkout
belongs to, so the command works from any directory:

```
sh install.sh
sh uninstall.sh
```

The compiled-in path is only a fallback: an explicit `-root` wins, then a
workspace found above the current directory, then the installer's path. The
scripts are described for users in the workspace README.

## Layout

```
skill-selector-app/
├── go.mod                      module github.com/maxx/skill-selector-app
├── .gitignore
├── main.go                     entry point: finds the workspace, wires packages together
├── install.sh                  build to $BIN_DIR as `skills` with the workspace path compiled in
├── uninstall.sh                remove that binary
└── internal/
    ├── ui/                     Bubble Tea TUI (model, update, view)
    │   ├── doc.go
    │   ├── model.go            root Model, Init/Update/View
    │   ├── layout.go           window frame: padding, footer rows, side-by-side panels
    │   ├── keys.go             key bindings; helpFor picks the footer hints per focused panel
    │   ├── theme/
    │   │   └── theme.go        shared chrome colours (borders, title tab)
    │   └── component/          independent presentation components, one per folder
    │       ├── doc.go          the component rules
    │       ├── panel/
    │       │   └── panel.go    titled sharp-cornered box: New(title, w, h), View(content)
    │       ├── list/
    │       │   └── list.go     scrollable plain list with cursor
    │       ├── checklist/
    │       │   └── checklist.go scrollable checked/clear list with cursor
    │       ├── searchbar/
    │       │   └── searchbar.go single-line text field with block caret
    │       ├── fileview/
    │       │   └── fileview.go word-wrapped, vertically scrolling text viewer
    │       └── markdown/
    │           └── markdown.go Glamour renderer: markdown in, styled text out
    ├── fuzzy/                  pure utility: Match predicate and generic Filter
    │   └── fuzzy.go
    ├── app/                    general business logic, coordinates skills + files
    │   ├── doc.go
    │   └── app.go              App: libraries, skills of the current library, toggle links
    ├── skills/                 describes the workspace layout
    │   ├── doc.go
    │   └── skill.go            Repo: FindRoot, ExposedDir, LibrariesDir, LibrarySkillsDir, SkillFile
    └── files/                  filesystem access only, no business logic
        ├── doc.go
        ├── read.go             Read(path)
        ├── symlink.go          IsLinked, Link, Unlink, Blocker
        └── list.go             ListDirs(path)
```

## Packages

### `main` (module root)

The `main` package, in `main.go` at the module root so `go build` here
produces the executable. It locates the agents workspace, constructs the
`app.App`, builds the `ui.Model` and starts the Bubble Tea program. The
workspace root is taken from the `-root` flag, else found with
`skills.FindRoot` by walking up from the current directory, else taken from
`defaultRoot`, a variable the installer sets with `-ldflags -X` so the
installed binary can be run from anywhere.

### `internal/ui`

Owns everything Bubble Tea: the root model, `Init`, `Update`, `View`, key
handling and rendering. It never makes decisions about skills or files; it
calls into `app` and displays the result.

The visual chrome is a port of the grove TUI (`~/Experiments/grove`) and is
split into two subpackages so screens stay free of styling detail:

- `ui/theme` holds the colour palette. The active border and title tab are
  ANSI 111 (light blue), inactive borders are ANSI 60 (slate), tab text is
  ANSI 235 (near black), and the list cursor row is a bar of `#bac2de`
  behind the terminal's default text colour.
- `ui/component/panel` draws a lazygit-style pane: square corners, the title
  rendered as a bold highlighted tab embedded in the top border (`┌─ Title ──┐`),
  content padded one column from each side, always filled to the exact
  width and height requested. `panel.Tab` renders the standalone app title in
  the same style.
- `ui/component/list` is a scrolling list of `Item{ID, Label}` rows. The
  cursor row is a full-width bar: bold default-colour text on a
  `theme.CursorBg` background. It owns the cursor and scroll offset, wraps at both ends,
  and keeps the cursor on the same ID across `SetItems` or snaps to the top.
  `ScrollUp` and `ScrollDown` move the window without the cursor, stopping at
  the ends; the next cursor move brings the cursor back into view. `SetSize`
  only re-anchors the window on the cursor when the height actually changes,
  so the per-render re-size does not undo a wheel scroll.
- `ui/component/checklist` is a scrolling list of
  `Item{ID, Checked, Disabled, Label}` rows. Unchecked rows show U+E640,
  checked rows U+F4A7 (both Nerd Font glyphs), the cursor row is a
  full-width bar of `theme.CursorBg` behind bold default-colour text. A
  disabled row is drawn in `theme.TextInactive` to show it cannot currently
  be checked; the cursor still lands on it and the glyph still follows
  `Checked`. Each row is rendered with one combined style so the dim
  foreground and the cursor bar do not reset each other mid-line. The
  component owns the cursor and scroll offset and wraps at both ends;
  checked and disabled state are passed in by the caller and never mutated
  by the component. `Items` hands back a copy of the rows as last set so the
  caller can act on exactly what is shown. Scrolling without the cursor
  works as in `list`.

- `ui/component/searchbar` is a single-line text field on a dark background
  with a block caret, ported from grove's text input. It owns the typed value
  and caret position only. Printable characters, space, backspace, delete,
  arrows, home and end are consumed; enter, up, down and tab are left for the
  caller. It knows nothing about what the text is used for.

- `ui/component/fileview` shows a block of text in a fixed area. Lines are
  soft-wrapped to the width with lipgloss, and the view scrolls vertically by
  wrapped line, stopping at both ends rather than wrapping around. It owns
  the scroll offset and wrapped lines only.

- `ui/component/markdown` is a stateless renderer built on Glamour. `Render`
  takes markdown and a width and returns ANSI-styled text wrapped to that
  width. A leading YAML frontmatter block is split off and drawn as a header
  (name in the accent colour, description muted) instead of literal dashes
  and key/value lines. The style starts from Glamour's dark theme with the
  heading, rule and bullet colours taken from `ui/theme` and the document
  margin removed, since the panel already pads its interior. Code fences are
  syntax-highlighted by Chroma through Glamour.

The Details panel shows the `SKILL.md` of the skill under the Skills cursor
via `app.SkillFile`. The source is piped through `markdown.Render` at the
panel's inner width and the result handed to `fileview`; the two components
never see each other. Because Glamour wraps at render time, the root model
re-renders when the panel width changes rather than letting the viewer
re-wrap styled text. Arrows scroll it while the panel has focus. When no
skill is under the cursor the panel shows a muted "No skill selected".

#### Panel pipeline

The three panels form a chain: the library under the left cursor decides
which skills are listed, and the skill under the middle cursor decides which
file is shown. The components never see each other; the root model joins
them in three sync functions that each call the next:

```
syncLibraries: app.Libraries() → list.Item → fuzzy.Filter(libSearch) → list.SetItems → syncLibrary
syncLibrary:   library under cursor → app.SelectLibrary → syncSkills
syncSkills:    app.Skills() → checklist.Item{Checked: app.IsSelected} → fuzzy.Filter(search) → checklist.SetItems → syncDetails
syncDetails:   skill under cursor → app.SkillFile → markdown.Render → fileview.SetContent
```

`syncLibrary` is a no-op while the cursor stays on the same library, and
`syncDetails` only re-reads and rewinds when the cursor lands on a different
skill, so a toggle or a re-filter that leaves the cursor in place keeps the
reader's scroll position. Because skill IDs are directory paths, none of the
old library's IDs survive a change between two installed libraries, so the
checklist snaps to its first row: the first skill of a newly chosen library
is selected automatically and its file appears in Details. Moving between an
installed library and the built-in All library is the exception: the skill
under the cursor exists in both lists, so the cursor stays on it.

While All is current, `syncSkills` labels each row `name (library)` via
`skillLabel`, since libraries can share skill names. The fuzzy filter runs
over that whole label, so typing a library name narrows the list to its
skills. Outside All the label is the bare skill name.

`internal/fuzzy` is a pure utility with no UI or app knowledge. `Match` is a
case-insensitive subsequence test that ignores spaces in the query, so
"code rev" matches "code-review". `Filter` is a generic order-preserving
filter over any slice.

Selection behaviour across filter changes is owned by the list components:
`SetItems` keeps the cursor on the row with the same ID when it is still
present, and snaps to the top row (scrolling back to the start) when it has
been filtered out.

Left and right are matched before any panel sees a key and switch panel.
That means those arrows do not move the caret inside a search bar; home and
end still do. Up and down answer to the up and down arrows and to shift+tab
and tab, all of which are matched before the search bars too.

Key routing while the Skill Libraries panel has focus: up and down move the
cursor and run `syncLibrary`; every other key goes to that panel's search
bar, and a consumed key re-runs `syncLibraries`. While the Skills panel has
focus: up and down move the cursor and run `syncDetails`, enter toggles the
row under the cursor via `app.ToggleSkill` and re-runs `syncSkills`, ctrl+a
toggles every row currently shown, and every other key goes to the search
bar.

Toggle all works on the rows the checklist is showing, so a search narrows
its reach. Shadowed rows, those whose name is already held in the exposed
directory by another library's skill or a foreign entry, count as toggled:
they are left out of both the decision and the sweep. If any shown skill is
unselected and not shadowed, all such skills are selected via
`app.SetSelected`; if every shown skill is selected or shadowed, all are
deselected. While selecting, a skill whose name was taken earlier in the same
sweep (the second `tdd` in the All view, say) is skipped silently, since it is
shadowed by the time its turn comes. Any other skill that cannot be changed
is skipped rather than aborting the sweep, and the failures are summarised in
the footer. It is bound to ctrl+a because
ctrl+enter is indistinguishable from enter in a terminal: both arrive as a
carriage return unless the terminal speaks the kitty keyboard protocol,
which Bubble Tea v1 does not parse. The binding is only matched while Skills
has focus, so ctrl+a still means home in the Skill Libraries search. Up and down also answer to k and j
in every panel; they are matched before the search bar sees the key, so
those two letters cannot be typed into a search. All other letters and space
type into the search, which is why toggle is enter rather than space. Quit
is esc or ctrl+c.

When `app.ToggleSkill` or `app.SelectLibrary` returns an error, its message
is shown in the footer row in `theme.Error` in place of the help until the
next key. The usual cause is a name clash: two libraries each have a skill
of the same name and one is already exposed. The message names the library
that holds the name (`exposing tdd: already exposed from pstack`), so the
user knows which link to turn off first. Shadowed rows are passed to the
checklist as `Disabled` and drawn dimmed, and `skillLabel` appends
`[skill with same name already exists]` after the name (and after the
library suffix in the All view), so a clash is visible before any key is
pressed. The fuzzy filter sees the suffix like the rest of the label.

Both search panels share one layout helper, `searchPanelContent`, that
stacks the top padding, the labelled search row, a blank row and the list
body. Each search bar is focused only while its panel is.

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
split 1:1:2 with a single blank column between each. Left and right arrows
cycle focus through the three panels, wrapping at both ends. Tab and
shift+tab are aliases for down and up in every panel. ctrl+tab is not bound
because a terminal sends the same byte for it as for plain tab. A left click inside a panel focuses it; the focused
panel takes the active border colour. A left click inside the panel that
already has focus acts on the row under the pointer: in Skill Libraries it
selects that library; in Skills it selects that skill, and a click on the
skill that is already selected toggles its link, as enter would. So from
another panel, a skill is toggled with three clicks: focus, select, toggle.
`listRowAt` maps the
window row to a list row from the shared search-panel layout, and clicks on
the search bar, the padding or below the last row do nothing. The mouse wheel scrolls whichever
panel is under the pointer, one row per notch, regardless of focus: the
library and skill lists scroll their window without moving the cursor, and
the details file scrolls by wrapped line. `panelGeometry` is the single
source of the panel widths and height, used both to draw them and, in
`panelAt`, to map a click or wheel event to a panel, so the two cannot drift
apart. Mouse reporting is enabled in `main` with `tea.WithMouseCellMotion`.
Mouse use is not listed in the help footer. The footer is rendered with
`bubbles/help` in short mode from `keyMap.helpFor(focused)` in `keys.go`, so
every panel lists up, down, next panel, previous panel and quit, and Skills
alone appends toggle and toggle all at the end.

### `internal/app`

General business logic. It is the only package that joins `skills` (where
things live) with `files` (how to touch them), and it exposes the operations
the UI needs. It knows nothing about the terminal and can be driven from
tests or a future CLI without changes.

- `New(repo)` lists the subdirectories of `repo.LibrariesDir()` once; each
  is a `Library` whose ID is its path and whose label is its name. The list
  always starts with the built-in combined library, labelled
  `All [built-in]`, whose ID is the constant `AllLibraryID` (`*`, which can
  never be a directory path). It is first so a fresh start shows every skill.
- `SelectLibrary(id)` lists the subdirectories of that library's skills
  directory; each is a `Skill` whose ID is its path and whose `Library`
  field names the library it came from. An unknown or empty ID clears the
  list. A library with no skills directory yields an empty list and an error
  for the UI to report. Selecting `AllLibraryID` loads every installed
  library's skills, sorted by skill name and then library name so same-named
  skills sit together; a library that cannot be listed is skipped, the others
  still load, and the failures come back as one error.
- `IsSelected(id)` asks `files.IsLinked` whether the skill is linked into
  `repo.ExposedDir()`. It is read from disk on every call, so links made or
  removed outside the program are reflected immediately.
- `ShadowedBy(id)` asks `files.Blocker` what holds the skill's name in the
  exposed directory when it is neither free nor the skill's own link. When
  the blocker is a skill of an installed library it returns that library's
  name; for anything else it returns the blocker's path; "" means the skill
  can be exposed or already is. Like `IsSelected` it reads from disk each
  call.
- `SetSelected(id, selected)` links the skill into the exposed directory or
  unlinks it, and is a no-op when it is already in that state. Before
  linking it checks `ShadowedBy` and fails with a message naming what is in
  the way: `already exposed from <library>` or
  `name already taken by <path>`. The existing entry is never replaced.
- `ToggleSkill(id)` flips the skill's state via `SetSelected`.
- `SkillFile(id)` reads the skill's `SKILL.md`. If it cannot be read the
  returned text says so, so the details panel always has content.

Selection state is not stored anywhere: the symlinks in the exposed
directory are the state.

### `internal/skills`

Describes the layout of the agents workspace and nothing more. `Repo{Root}`
builds every path the program needs:

| Method | Path |
|---|---|
| `ExposedDir()` | `<root>/skills`, the directory the agent reads; selected skills are linked here |
| `LibrariesDir()` | `<root>/skill-libraries`, one cloned library per subdirectory |
| `LibrarySkillsDir(lib)` | `<lib>/skills`, the flat directory of skills in a library |
| `SkillFile(skill)` | `<skill>/SKILL.md` |

`FindRoot(start)` walks up from `start` to the first directory containing
both `skills/` and `skill-selector-app/`. Those are the markers because they
are committed; `skill-libraries/` is gitignored and absent on a fresh
clone. It uses `files.ListDirs` for the check so the package still touches
the filesystem only through `files`.

### `internal/files`

Thin wrapper over filesystem access. Its exported API is exactly six
functions and nothing else; helpers stay unexported so callers never depend
on how a check is performed:

- `Read(path)` returns a file's contents as a string, ready to hand to the
  `fileview` component.
- `ListDirs(path)` returns the full path of every directory directly beneath
  `path`, sorted by name. Symlinks that resolve to a directory count;
  files and dangling links do not.
- `IsLinked(dir, target)` reports whether `dir` holds a symlink to the
  folder `target`.
- `Link(dir, target)` creates that symlink, storing `target`'s absolute path.
- `Unlink(dir, target)` removes it.
- `Blocker(dir, target)` reports what stops `Link` from succeeding: "" when
  the name is free or already links to `target`, the destination of a
  symlink to somewhere else, or the path of a real file or directory that
  holds the name.

Every symlink function uses the same convention: the link lives in `dir` and
is named after the last path element of `target`, so linking
`/lib/skills/code-review` into `/exposed` creates `/exposed/code-review`.
Paths are compared after resolving to their canonical absolute form, so a
relative `target` and its absolute spelling are treated as the same folder.

The package is deliberately conservative about deleting. `Unlink` only
removes an entry that is a symlink pointing at `target`; a real directory of
the same name, or a link to somewhere else, is left in place and reported as
an error. Unlinking something that is not there is a no-op. `Link` fails
rather than overwriting an existing entry. Symlink checks use `os.Lstat` so a
link to a directory is seen as a link, not as a directory.

No business logic lives here.

## Dependency direction

```
     main
       │
       ▼
      ui ──▶ app ──▶ skills ──▶ files
              │                  ▲
              └──────────────────┘
```

Dependencies only point downward. `files` imports nothing from the project,
`skills` imports `files`, `app` imports `skills` and `files`, and `ui`
imports only `app`. `main` imports `skills` as well as `app` and `ui`
because it builds the `skills.Repo` the app is given. Nothing
imports `ui`.

## Conventions

- Packages live under `internal/` so they cannot be imported by other modules.
- Each package has a `doc.go` stating its responsibility and what it must not
  do.
- Third-party dependencies are added when first used. The `ui` package pulls
  in `charmbracelet/bubbletea`, `charmbracelet/lipgloss`,
  `charmbracelet/bubbles` and `charmbracelet/glamour`; nothing outside `ui`
  may import them.
