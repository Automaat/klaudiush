// Package mdtable provides markdown table formatting that passes markdownlint validation.
package mdtable

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// Alignment specifies column alignment.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// WidthMode specifies how column widths are calculated.
type WidthMode int

const (
	// WidthModeDisplay uses display width (proper for CJK/emoji).
	// Tables will be visually aligned but may fail markdownlint MD060.
	WidthModeDisplay WidthMode = iota

	// WidthModeByte uses byte length for width calculations.
	// Tables will pass markdownlint MD060 but may not be visually aligned for Unicode.
	WidthModeByte
)

const (
	// minSeparatorWidth is the minimum width for table separator dashes (---).
	minSeparatorWidth = 3

	// cellPadding is the number of spaces added around cell content.
	cellPadding = 2

	// halfDivisor is used for calculating center alignment padding.
	halfDivisor = 2
)

// Table represents a markdown table.
type Table struct {
	headers    []string
	alignments []Alignment
	rows       [][]string
	widthMode  WidthMode
}

// New creates a new Table with the given headers.
func New(headers ...string) *Table {
	alignments := make([]Alignment, len(headers))
	for i := range alignments {
		alignments[i] = AlignLeft
	}

	return &Table{
		headers:    headers,
		alignments: alignments,
		rows:       make([][]string, 0),
		widthMode:  WidthModeDisplay, // Default: proper display width
	}
}

// SetWidthMode sets the width calculation mode.
func (t *Table) SetWidthMode(mode WidthMode) *Table {
	t.widthMode = mode

	return t
}

// SetAlignment sets the alignment for a column.
func (t *Table) SetAlignment(col int, align Alignment) *Table {
	if col >= 0 && col < len(t.alignments) {
		t.alignments[col] = align
	}

	return t
}

// SetAlignments sets all column alignments.
func (t *Table) SetAlignments(alignments ...Alignment) *Table {
	for i, align := range alignments {
		if i < len(t.alignments) {
			t.alignments[i] = align
		}
	}

	return t
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cells ...string) *Table {
	// Normalize row to match header count
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cells) {
			row[i] = cells[i]
		}
	}

	t.rows = append(t.rows, row)

	return t
}

// String renders the table as a markdown string.
func (t *Table) String() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := t.calculateWidths()

	var sb strings.Builder

	// Render header row
	sb.WriteString(t.renderRow(t.headers, widths))
	sb.WriteString("\n")

	// Render separator row
	sb.WriteString(t.renderSeparator(widths))
	sb.WriteString("\n")

	// Render data rows
	for _, row := range t.rows {
		sb.WriteString(t.renderRow(row, widths))
		sb.WriteString("\n")
	}

	return sb.String()
}

// calculateWidths determines the width of each column.
func (t *Table) calculateWidths() []int {
	widths := make([]int, len(t.headers))
	getWidth := widthFunc(t.widthMode)

	// Start with header widths
	for i, header := range t.headers {
		widths[i] = getWidth(sanitize(header))
	}

	// Check data rows
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				w := getWidth(sanitize(cell))
				if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	// Minimum width for separator (---)
	for i := range widths {
		if widths[i] < minSeparatorWidth {
			widths[i] = minSeparatorWidth
		}
	}

	return widths
}

// renderRow renders a single row with proper padding.
func (t *Table) renderRow(cells []string, widths []int) string {
	var sb strings.Builder

	sb.WriteString("|")

	for i, cell := range cells {
		if i >= len(widths) {
			break
		}

		sanitized := sanitize(cell)
		padded := padWithMode(sanitized, widths[i], t.alignments[i], t.widthMode)

		sb.WriteString(" ")
		sb.WriteString(padded)
		sb.WriteString(" |")
	}

	return sb.String()
}

// renderSeparator renders the separator row between header and data.
// Format: |:-----|-----| (no spaces around dashes).
func (t *Table) renderSeparator(widths []int) string {
	var sb strings.Builder

	sb.WriteString("|")

	for i, w := range widths {
		// Add cellPadding to width to account for the spaces in data cells
		totalWidth := w + cellPadding

		switch t.alignments[i] {
		case AlignLeft:
			sb.WriteString(":")
			sb.WriteString(strings.Repeat("-", totalWidth-1))
		case AlignCenter:
			sb.WriteString(":")
			sb.WriteString(strings.Repeat("-", totalWidth-cellPadding))
			sb.WriteString(":")
		case AlignRight:
			sb.WriteString(strings.Repeat("-", totalWidth-1))
			sb.WriteString(":")
		}

		sb.WriteString("|")
	}

	return sb.String()
}

// sanitize cleans a cell value for markdown table use.
func sanitize(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)

	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")

	// Escape unescaped pipe characters (avoid double-escaping)
	s = escapePipes(s)

	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	return s
}

// escapePipes escapes pipe characters that aren't already escaped.
func escapePipes(s string) string {
	var result strings.Builder

	runes := []rune(s)

	for i := range runes {
		if runes[i] == '|' {
			// Check if already escaped (preceded by backslash)
			if i > 0 && runes[i-1] == '\\' {
				// Already escaped, just add the pipe
				result.WriteRune('|')
			} else {
				// Not escaped, add escape
				result.WriteString("\\|")
			}
		} else {
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

// displayWidth returns the display width of a string.
// Uses runewidth to correctly handle East Asian characters and emoji
// which may have a display width of 2.
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// byteWidth returns the byte length of a string.
// This is used for markdownlint MD060 compliance, which uses byte positions.
func byteWidth(s string) int {
	return len(s)
}

// widthFunc returns the appropriate width calculation function for the mode.
func widthFunc(mode WidthMode) func(string) int {
	if mode == WidthModeByte {
		return byteWidth
	}

	return displayWidth
}

// padWithMode pads a string to the given width with the specified alignment and width mode.
func padWithMode(s string, width int, align Alignment, mode WidthMode) string {
	getWidth := widthFunc(mode)
	w := getWidth(s)

	if w >= width {
		return s
	}

	padding := width - w

	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + s
	case AlignCenter:
		left := padding / halfDivisor
		right := padding - left

		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	default: // AlignLeft
		return s + strings.Repeat(" ", padding)
	}
}

// Format formats raw table data into a valid markdown table.
// Headers is the first row, followed by data rows.
func Format(headers []string, rows [][]string, alignments ...Alignment) string {
	return FormatWithMode(headers, rows, WidthModeDisplay, alignments...)
}

// FormatWithMode formats raw table data with a specific width calculation mode.
func FormatWithMode(
	headers []string,
	rows [][]string,
	mode WidthMode,
	alignments ...Alignment,
) string {
	t := New(headers...)
	t.SetWidthMode(mode)
	t.SetAlignments(alignments...)

	for _, row := range rows {
		t.AddRow(row...)
	}

	return t.String()
}

// FormatSimple creates a simple left-aligned table.
func FormatSimple(headers []string, rows [][]string) string {
	return Format(headers, rows)
}
