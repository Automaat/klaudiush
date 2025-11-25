// Package validators provides shared markdown validation utilities
package validators

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/smykla-labs/klaudiush/pkg/mdtable"
)

const (
	maxTruncateLength = 60
)

// MarkdownState represents the parsing state at a given position
type MarkdownState struct {
	InCodeBlock bool
	StartLine   int  // 0-indexed line number where this state begins (0 = start of file)
	EndsAtEOF   bool // true if this fragment includes the last line of the file
	// Future: InComment, ListDepth, etc.
}

// MarkdownAnalysisResult contains markdown validation warnings
type MarkdownAnalysisResult struct {
	Warnings       []string
	TableSuggested map[int]string // Line number -> suggested formatted table
}

// AnalysisOptions contains options for markdown analysis.
type AnalysisOptions struct {
	// CheckTableFormatting enables table formatting validation.
	// Default: true
	CheckTableFormatting bool

	// TableWidthMode controls how table column widths are calculated.
	// Default: mdtable.WidthModeDisplay
	TableWidthMode mdtable.WidthMode
}

// DefaultAnalysisOptions returns the default analysis options.
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		CheckTableFormatting: true,
		TableWidthMode:       mdtable.WidthModeDisplay,
	}
}

// listContext tracks the context of a list item for indentation validation
type listContext struct {
	lineNum           int
	indent            int
	sawEmptyLineAfter bool
}

var (
	codeBlockRegex = regexp.MustCompile(`^[[:space:]]*` + "```")
	listItemRegex  = regexp.MustCompile(
		`^[[:space:]]*[-*+][[:space:]]|^[[:space:]]*[0-9]+\.[[:space:]]`,
	)
	headerRegex    = regexp.MustCompile(`^#+[[:space:]]`)
	commentRegex   = regexp.MustCompile(`^<!--`)
	emptyLineRegex = regexp.MustCompile(`^[[:space:]]*$`)
)

// DetectMarkdownState scans content up to a given line to determine the state.
// This allows fragment validation to start with the correct context.
func DetectMarkdownState(content string, upToLine int) MarkdownState {
	state := MarkdownState{InCodeBlock: false}

	if upToLine <= 0 {
		return state
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	for scanner.Scan() && lineNum < upToLine {
		line := scanner.Text()
		lineNum++

		if isCodeBlockMarker(line) {
			state.InCodeBlock = !state.InCodeBlock
		}
	}

	return state
}

// AnalyzeMarkdown performs line-by-line markdown analysis and returns warnings.
// If initialState is provided, it uses that as the starting state (for fragment validation).
// Options can be provided to control table formatting validation.
func AnalyzeMarkdown(
	content string,
	initialState *MarkdownState,
	opts ...AnalysisOptions,
) MarkdownAnalysisResult {
	result := MarkdownAnalysisResult{
		Warnings:       []string{},
		TableSuggested: make(map[int]string),
	}

	if content == "" {
		return result
	}

	// Use provided options or defaults
	options := DefaultAnalysisOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Check for table issues and collect suggestions if enabled
	if options.CheckTableFormatting {
		checkTables(content, &result, options.TableWidthMode)
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	prevLine := ""
	prevPrevLine := ""

	// Use initial state if provided, otherwise start fresh
	inCodeBlock := false
	if initialState != nil {
		inCodeBlock = initialState.InCodeBlock
	}

	var lastList *listContext

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Track list context for indentation validation
		switch {
		case isListItem(line):
			lastList = &listContext{
				lineNum:           lineNum,
				indent:            getListIndent(line),
				sawEmptyLineAfter: false,
			}
		case lastList != nil && isEmptyLine(line):
			lastList.sawEmptyLineAfter = true
		case lastList != nil && !isEmptyLine(line) && !isListItem(line) && !isCodeBlockMarker(line):
			// Reset list context if we encounter non-list, non-empty, non-code content
			lastList = nil
		}

		// Check for code block markers and indentation
		if isCodeBlockMarker(line) {
			checkCodeBlockIndentation(line, lastList, lineNum, &result.Warnings)
			checkMultipleEmptyLinesBeforeCodeBlock(
				prevLine,
				prevPrevLine,
				lineNum,
				&result.Warnings,
			)
			inCodeBlock = checkCodeBlock(line, prevLine, lineNum, inCodeBlock, &result.Warnings)
			// Reset list context after code block
			if !inCodeBlock {
				lastList = nil
			}
		} else {
			inCodeBlock = checkCodeBlock(line, prevLine, lineNum, inCodeBlock, &result.Warnings)
		}

		// Skip list checks inside code blocks
		if inCodeBlock {
			prevPrevLine = prevLine
			prevLine = line

			continue
		}

		// Skip validation for first line (can't check previous line)
		if lineNum > 1 {
			// Check for first list item (transition from non-list to list)
			checkListItem(line, prevLine, lineNum, &result.Warnings)

			// Check for content immediately after headers
			checkHeader(line, prevLine, lineNum, &result.Warnings)
		}

		prevPrevLine = prevLine
		prevLine = line
	}

	return result
}

// checkCodeBlock checks for code block markers and validates spacing
func checkCodeBlock(line, prevLine string, lineNum int, inCodeBlock bool, warnings *[]string) bool {
	if !isCodeBlockMarker(line) {
		return inCodeBlock
	}

	if !inCodeBlock {
		// Opening code block
		if !isEmptyLine(prevLine) && prevLine != "" {
			*warnings = append(*warnings,
				fmt.Sprintf("⚠️  Line %d: Code block should have empty line before it", lineNum),
				fmt.Sprintf("   Previous line: '%s'", truncate(prevLine)),
			)
		}

		return true
	}

	// Closing code block
	return false
}

// checkCodeBlockIndentation validates code block indentation within list items
func checkCodeBlockIndentation(
	line string,
	lastList *listContext,
	lineNum int,
	warnings *[]string,
) {
	if lastList == nil || !lastList.sawEmptyLineAfter {
		return
	}

	indent := getIndentation(line)

	// If code block has no indentation at all, it's a separate block, not part of the list
	// Only warn if it has some indentation but not enough (partial indentation suggests
	// it was intended to be part of the list)
	if indent > 0 && indent < lastList.indent {
		*warnings = append(
			*warnings,
			fmt.Sprintf(
				"⚠️  Line %d: Code block in list item should be indented by at least %d spaces",
				lineNum,
				lastList.indent,
			),
			fmt.Sprintf(
				"   Found: %d spaces, expected: at least %d spaces",
				indent,
				lastList.indent,
			),
		)
	}
}

// checkMultipleEmptyLinesBeforeCodeBlock validates that there's only one empty line before code blocks
func checkMultipleEmptyLinesBeforeCodeBlock(
	prevLine, prevPrevLine string,
	lineNum int,
	warnings *[]string,
) {
	// Check if we have two consecutive empty lines before the code block
	// lineNum > 3 ensures we have at least 3 lines processed, so prevPrevLine is from actual content
	if lineNum > 3 && isEmptyLine(prevLine) && isEmptyLine(prevPrevLine) {
		*warnings = append(
			*warnings,
			fmt.Sprintf(
				"⚠️  Line %d: Code block should have only one empty line before it, not multiple",
				lineNum,
			),
			"   Found multiple consecutive empty lines before code block",
		)
	}
}

// checkListItem validates list item spacing
func checkListItem(line, prevLine string, lineNum int, warnings *[]string) {
	if !isListItem(line) {
		return
	}

	if shouldWarnAboutListSpacing(prevLine) {
		*warnings = append(*warnings,
			fmt.Sprintf("⚠️  Line %d: First list item should have empty line before it", lineNum),
			fmt.Sprintf("   Previous line: '%s'", truncate(prevLine)),
		)
	}
}

// shouldWarnAboutListSpacing determines if a list item needs spacing before it
func shouldWarnAboutListSpacing(prevLine string) bool {
	return !isEmptyLine(prevLine) &&
		prevLine != "" &&
		!isListItem(prevLine) &&
		!isHeader(prevLine)
}

// checkHeader validates header spacing
func checkHeader(line, prevLine string, lineNum int, warnings *[]string) {
	if !isHeader(prevLine) {
		return
	}

	// Lists are allowed directly after headers
	if shouldWarnAboutHeaderSpacing(line) {
		*warnings = append(*warnings,
			fmt.Sprintf("⚠️  Line %d: Header should have empty line after it", lineNum-1),
			fmt.Sprintf("   Header: '%s'", truncate(prevLine)),
			fmt.Sprintf("   Next line: '%s'", truncate(line)),
		)
	}
}

// shouldWarnAboutHeaderSpacing determines if content after a header needs spacing
func shouldWarnAboutHeaderSpacing(line string) bool {
	return line != "" &&
		!isEmptyLine(line) &&
		!isHeader(line) &&
		!isComment(line) &&
		!isListItem(line)
}

// getListIndent calculates the required indentation for list item content
// For "4. text" → returns 3 (length of "4. ")
// For "- text" → returns 2 (length of "- ")
// For "  - text" → returns 4 (length of "  - ")
func getListIndent(line string) int {
	re := regexp.MustCompile(`^([[:space:]]*)([-*+]|[0-9]+\.)[[:space:]]`)
	matches := re.FindStringSubmatch(line)

	const minRequiredMatches = 3 // Full match + 2 capture groups

	if len(matches) < minRequiredMatches {
		return 0
	}

	leadingSpace := matches[1]
	marker := matches[2]

	return len(leadingSpace) + len(marker) + 1 // +1 for space after marker
}

// getIndentation returns the number of leading spaces in a line
func getIndentation(line string) int {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return i
		}
	}

	return len(line)
}

// isCodeBlockMarker checks if line starts a code block
func isCodeBlockMarker(line string) bool {
	return codeBlockRegex.MatchString(line)
}

// isListItem checks if line is a list item
func isListItem(line string) bool {
	return listItemRegex.MatchString(line)
}

// isHeader checks if line is a header
func isHeader(line string) bool {
	return headerRegex.MatchString(line)
}

// isComment checks if line is an HTML comment
func isComment(line string) bool {
	return commentRegex.MatchString(line)
}

// isEmptyLine checks if line is empty or whitespace-only
func isEmptyLine(line string) bool {
	return emptyLineRegex.MatchString(line)
}

// truncate truncates string to maxTruncateLength
func truncate(s string) string {
	if len(s) <= maxTruncateLength {
		return s
	}

	return s[:maxTruncateLength]
}

// checkTables parses markdown tables and checks for formatting issues.
// When issues are found, it adds warnings and suggests properly formatted tables.
func checkTables(content string, result *MarkdownAnalysisResult, widthMode mdtable.WidthMode) {
	parseResult := mdtable.Parse(content)

	for _, table := range parseResult.Tables {
		// Check if the table needs reformatting by comparing original vs formatted
		formatted := mdtable.FormatTableWithMode(&table, widthMode)
		original := strings.Join(table.RawLines, "\n") + "\n"

		if formatted != original {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("⚠️  Line %d: Markdown table has formatting issues", table.StartLine),
				"   Table should be properly formatted with consistent column widths",
			)
			result.TableSuggested[table.StartLine] = formatted
		}
	}

	// Add any specific issues from parsing
	for _, issue := range parseResult.Issues {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("⚠️  Line %d: %s", issue.Line, issue.Message),
		)
	}
}
