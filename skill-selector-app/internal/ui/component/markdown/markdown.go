// Package markdown turns markdown source into ANSI-styled terminal text at a
// given width, using Glamour with a style derived from the theme colours. It
// is stateless: markdown in, styled string out. Displaying or scrolling the
// result is another component's job.
package markdown

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/maxx/skill-selector-app/internal/ui/theme"
)

// Frontmatter is the parsed YAML header of a skill file. Only the keys shown
// in the rendered header are extracted; other keys are ignored.
type Frontmatter struct {
	Name        string
	Description string
}

// Render converts src to styled text wrapped at width. A leading YAML
// frontmatter block is stripped and rendered as a header of the name and
// description instead of as literal dashes and key/value lines. On a
// rendering error the raw source is returned so the caller always has
// something to show.
func Render(src string, width int) string {
	fm, body := SplitFrontmatter(src)
	header := renderHeader(fm, width)

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(styleConfig()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return header + body
	}
	out, err := r.Render(body)
	if err != nil {
		return header + body
	}
	return header + strings.Trim(out, "\n")
}

// SplitFrontmatter separates a leading "---" delimited YAML block from the
// rest of the document. Without one, the frontmatter is empty and body is
// src unchanged.
func SplitFrontmatter(src string) (Frontmatter, string) {
	lines := strings.Split(src, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Frontmatter{}, src
	}
	var fm Frontmatter
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return fm, strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\n")
		}
		key, value, ok := strings.Cut(lines[i], ":")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.TrimSpace(key) {
		case "name":
			fm.Name = value
		case "description":
			fm.Description = value
		}
	}
	// No closing delimiter: treat the whole thing as body.
	return Frontmatter{}, src
}

// renderHeader draws the frontmatter as a title line and a muted,
// wrapped description, followed by a blank line. Empty frontmatter renders
// nothing.
func renderHeader(fm Frontmatter, width int) string {
	if fm.Name == "" && fm.Description == "" {
		return ""
	}
	var b strings.Builder
	if fm.Name != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.BorderActive).Bold(true).Render(fm.Name))
		b.WriteString("\n")
	}
	if fm.Description != "" {
		desc := lipgloss.NewStyle().Foreground(theme.TextInactive)
		if width > 0 {
			desc = desc.Width(width)
		}
		b.WriteString(desc.Render(fm.Description))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// styleConfig starts from Glamour's dark style and pulls the accent colours
// from the theme so headings and rules match the rest of the chrome. The
// document margin is removed because the panel already pads its interior.
func styleConfig() ansi.StyleConfig {
	cfg := styles.DarkStyleConfig
	accent := string(theme.BorderActive)
	muted := string(theme.TextInactive)
	zero := uint(0)

	cfg.Document.Margin = &zero
	cfg.Document.BlockPrefix = ""
	cfg.Document.BlockSuffix = ""

	cfg.Heading.Color = &accent
	cfg.H1.Color = &accent
	cfg.H1.BackgroundColor = nil
	cfg.H1.Prefix = "# "
	cfg.H1.Suffix = ""

	cfg.HorizontalRule.Color = &muted
	cfg.BlockQuote.StylePrimitive.Color = &muted
	cfg.Item.Color = &accent

	cfg.CodeBlock.Margin = &zero
	return cfg
}
