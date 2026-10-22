package ui

import (
	"strings"

	"github.com/maxx/skill-selector-app/internal/ui/components/panel"
)

const (
	// horizontalPadding is the blank margin between the terminal edge and the
	// content on the left and right.
	horizontalPadding = 4
	// verticalPadding is the blank margin above and below the content.
	verticalPadding = 2
	// footerRows is reserved beneath the body for the help footer. The panels
	// extend all the way down to it.
	footerRows = 1
	// panelGap is the number of blank columns between side-by-side panels.
	panelGap = 1
)

// frame lays the body and footer inside the padded window. The body fills
// everything above the footer, which sits flush with the bottom of the
// content area.
func frame(body, footer string, width, height int) string {
	contentWidth := max(0, width-horizontalPadding*2)
	contentHeight := max(0, height-verticalPadding*2)
	bodyHeight := max(0, contentHeight-footerRows)

	lines := make([]string, 0, contentHeight)
	lines = append(lines, fitBody(body, contentWidth, bodyHeight)...)
	lines = append(lines, panel.Fit(footer, contentWidth))
	for len(lines) < contentHeight {
		lines = append(lines, panel.Fit("", contentWidth))
	}
	if len(lines) > contentHeight {
		lines = lines[:contentHeight]
	}

	padded := make([]string, 0, height)
	blankLine := strings.Repeat(" ", max(width, 0))
	for i := 0; i < verticalPadding; i++ {
		padded = append(padded, blankLine)
	}
	leftPad := strings.Repeat(" ", horizontalPadding)
	for _, line := range lines {
		padded = append(padded, leftPad+panel.Fit(line, contentWidth))
	}
	for len(padded) < height {
		padded = append(padded, blankLine)
	}
	if len(padded) > height {
		padded = padded[:height]
	}
	return strings.Join(padded, "\n")
}

// sideBySide joins pre-rendered panels of equal height left to right with
// panelGap blank columns between each. widths[i] is the width panels[i] was
// rendered at; every column but the last is padded to it so the next panel
// starts on a fixed column even when a row is short.
func sideBySide(panels []string, widths []int, height int) string {
	split := make([][]string, len(panels))
	for i, p := range panels {
		split[i] = strings.Split(p, "\n")
	}
	gap := strings.Repeat(" ", panelGap)
	rows := make([]string, 0, height)
	for i := 0; i < height; i++ {
		var row strings.Builder
		for j, lines := range split {
			cell := ""
			if i < len(lines) {
				cell = lines[i]
			}
			if j < len(split)-1 {
				cell = panel.Fit(cell, widths[j]) + gap
			}
			row.WriteString(cell)
		}
		rows = append(rows, row.String())
	}
	return strings.Join(rows, "\n")
}

// splitWidths divides total columns between len(ratios) panels in the given
// proportions, after reserving panelGap between each pair. Rounding remainder
// goes to the last panel so the row always sums to total.
func splitWidths(total int, ratios ...int) []int {
	usable := max(0, total-panelGap*(len(ratios)-1))
	sum := 0
	for _, r := range ratios {
		sum += r
	}
	widths := make([]int, len(ratios))
	used := 0
	for i, r := range ratios[:len(ratios)-1] {
		widths[i] = usable * r / sum
		used += widths[i]
	}
	widths[len(ratios)-1] = max(0, usable-used)
	return widths
}

func fitBody(content string, width, height int) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		result = append(result, panel.Fit(line, width))
	}
	return result
}

// contentWidthFor returns the width available to content after horizontal
// padding is removed from the window width.
func contentWidthFor(windowWidth int) int {
	return max(0, windowWidth-horizontalPadding*2)
}
