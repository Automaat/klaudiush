package secrets

import (
	"strings"
)

//go:generate mockgen -source=detector.go -destination=detector_mock.go -package=secrets

// Detector is the interface for secret detection implementations.
type Detector interface {
	// Detect scans content for secrets and returns all findings.
	Detect(content string) []Finding
}

// PatternDetector implements Detector using compiled regex patterns.
type PatternDetector struct {
	patterns []Pattern
}

// NewPatternDetector creates a new PatternDetector with the given patterns.
func NewPatternDetector(patterns []Pattern) *PatternDetector {
	return &PatternDetector{
		patterns: patterns,
	}
}

// NewDefaultPatternDetector creates a PatternDetector with default patterns.
func NewDefaultPatternDetector() *PatternDetector {
	return NewPatternDetector(DefaultPatterns())
}

// Detect scans content for secrets using all configured patterns.
func (d *PatternDetector) Detect(content string) []Finding {
	if content == "" {
		return nil
	}

	var findings []Finding

	lines := strings.Split(content, "\n")

	for i, pattern := range d.patterns {
		matches := d.patterns[i].Regex.FindAllStringIndex(content, -1)

		for _, match := range matches {
			line, col := d.getPosition(lines, content, match[0])

			findings = append(findings, Finding{
				Pattern: &d.patterns[i],
				Match:   content[match[0]:match[1]],
				Line:    line,
				Column:  col,
			})
		}

		_ = pattern // avoid unused variable lint error
	}

	return findings
}

// getPosition returns the 1-indexed line and column for a byte offset.
func (*PatternDetector) getPosition(lines []string, _ string, offset int) (line, column int) {
	pos := 0

	for lineIdx, lineContent := range lines {
		lineEnd := pos + len(lineContent)
		if lineIdx < len(lines)-1 {
			lineEnd++ // account for newline character
		}

		if offset < lineEnd {
			return lineIdx + 1, offset - pos + 1
		}

		pos = lineEnd
	}

	// Fallback: return last line
	return len(lines), offset - pos + 1
}

// AddPatterns adds additional patterns to the detector.
func (d *PatternDetector) AddPatterns(patterns ...Pattern) {
	d.patterns = append(d.patterns, patterns...)
}
