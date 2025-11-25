package file

import (
	"cmp"
	"strings"

	"github.com/smykla-labs/klaudiush/pkg/logger"
)

// ExtractEditFragment extracts the edit region with surrounding context lines.
// It finds the oldStr in content, replaces it with newStr, and returns a fragment
// containing the edit plus contextLines before and after for proper linting context.
func ExtractEditFragment(
	content string,
	oldStr string,
	newStr string,
	contextLines int,
	log logger.Logger,
) string {
	// Find the position of oldStr in content
	idx := strings.Index(content, oldStr)
	if idx == -1 {
		log.Debug("old_string not found in file content")
		return ""
	}

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Find which line contains the start of oldStr
	charCount := 0
	startLine := 0

	for i, line := range lines {
		if charCount+len(line)+1 > idx { // +1 for newline
			startLine = i
			break
		}

		charCount += len(line) + 1
	}

	// Find which line contains the end of oldStr
	endIdx := idx + len(oldStr)
	charCount = 0
	endLine := 0

	for i, line := range lines {
		charCount += len(line) + 1

		if charCount >= endIdx {
			endLine = i
			break
		}
	}

	// Extract lines with context
	contextStart := max(0, startLine-contextLines)
	contextEnd := cmp.Or(min(endLine+contextLines, len(lines)-1), len(lines)-1)

	// Build fragment with the edit applied
	fragmentLines := make([]string, 0, contextEnd-contextStart+1)

	for i := contextStart; i <= contextEnd; i++ {
		fragmentLines = append(fragmentLines, lines[i])
	}

	// Strip trailing empty lines to avoid false positives from files with trailing blank lines.
	// These trailing blanks, when combined with preamble context, can create consecutive
	// blank lines that trigger MD012 (no-multiple-blanks) errors.
	fragmentLines = trimTrailingEmptyLines(fragmentLines)

	// Apply the replacement to the fragment
	fragment := strings.Join(fragmentLines, "\n")
	fragment = strings.Replace(fragment, oldStr, newStr, 1)

	return fragment
}

// trimTrailingEmptyLines removes excess trailing empty strings from a slice.
// Keeps at most one trailing empty string (preserving normal trailing newline)
// but removes additional ones (blank lines) that would cause MD012 errors
// when combined with preamble context.
func trimTrailingEmptyLines(lines []string) []string {
	// Count trailing empty lines
	trailingCount := 0

	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			break
		}

		trailingCount++
	}

	// Keep at most one trailing empty line (normal trailing newline)
	if trailingCount > 1 {
		lines = lines[:len(lines)-(trailingCount-1)]
	}

	return lines
}

// getFragmentStartLine returns the line number where the fragment starts (0-indexed).
// This accounts for context lines added before the actual edit location.
func getFragmentStartLine(content, oldStr string, contextLines int) int {
	idx := strings.Index(content, oldStr)
	if idx == -1 {
		return 0
	}

	lines := strings.Split(content, "\n")
	charCount := 0
	startLine := 0

	for i, line := range lines {
		if charCount+len(line)+1 > idx { // +1 for newline
			startLine = i
			break
		}

		charCount += len(line) + 1
	}

	return max(0, startLine-contextLines)
}
