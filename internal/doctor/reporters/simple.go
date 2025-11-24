// Package reporters provides output formatting for doctor check results
package reporters

import (
	"fmt"
	"strings"

	"github.com/smykla-labs/klaudiush/internal/doctor"
)

const (
	// categoryParts is the number of parts when splitting by category separator
	categoryParts = 2
)

// SimpleReporter provides simple checklist-style output
type SimpleReporter struct{}

// NewSimpleReporter creates a new SimpleReporter
func NewSimpleReporter() *SimpleReporter {
	return &SimpleReporter{}
}

// Report outputs the results in a simple checklist format
func (*SimpleReporter) Report(results []doctor.CheckResult, verbose bool) {
	// Group results by category
	categoryMap := groupByCategory(results)

	// Print header
	fmt.Println("Checking klaudiush health...")
	fmt.Println()

	// Print results by category
	printCategories(categoryMap, verbose)

	// Print summary
	printSummary(results)
}

// groupByCategory groups results by their category
func groupByCategory(results []doctor.CheckResult) map[string][]doctor.CheckResult {
	categoryMap := make(map[string][]doctor.CheckResult)

	for _, result := range results {
		category := extractCategory(result.Name)
		categoryMap[category] = append(categoryMap[category], result)
	}

	return categoryMap
}

// extractCategory extracts the category from a check name
func extractCategory(name string) string {
	parts := strings.SplitN(name, ":", categoryParts)
	if len(parts) == categoryParts {
		return strings.TrimSpace(parts[0])
	}

	return "Other"
}

// extractCheckName extracts just the check name without category prefix
func extractCheckName(name string) string {
	parts := strings.SplitN(name, ":", categoryParts)

	if len(parts) == categoryParts {
		return strings.TrimSpace(parts[1])
	}

	return name
}

// printCategories prints results grouped by category
func printCategories(categoryMap map[string][]doctor.CheckResult, verbose bool) {
	for category, categoryResults := range categoryMap {
		fmt.Printf("%s:\n", category)

		for _, result := range categoryResults {
			printResult(result, verbose)
		}

		fmt.Println()
	}
}

// printResult prints a single check result
func printResult(result doctor.CheckResult, verbose bool) {
	icon := getStatusIcon(result)
	checkName := extractCheckName(result.Name)

	// Print status line
	fmt.Printf("  %s %s", icon, checkName)

	if result.Message != "" {
		fmt.Printf(" - %s", result.Message)
	}

	fmt.Println()

	// Print details in verbose mode
	if verbose && len(result.Details) > 0 {
		printDetails(result.Details)
	}

	// Print fix suggestion
	if result.HasFix() && result.Status == doctor.StatusFail {
		fmt.Printf("     → Run: klaudiush doctor --fix\n")
	}
}

// printDetails prints detail lines
func printDetails(details []string) {
	for _, detail := range details {
		fmt.Printf("     %s\n", detail)
	}
}

// printSummary prints the summary line
func printSummary(results []doctor.CheckResult) {
	errorCount, warningCount, passedCount := countResults(results)

	fmt.Printf("Summary: %d error(s), %d warning(s), %d passed\n",
		errorCount, warningCount, passedCount)
}

// getStatusIcon returns the appropriate icon for a check result
func getStatusIcon(result doctor.CheckResult) string {
	switch result.Status {
	case doctor.StatusPass:
		return "✅"
	case doctor.StatusFail:
		switch result.Severity {
		case doctor.SeverityError:
			return "❌"
		case doctor.SeverityWarning:
			return "⚠️"
		default:
			return "ℹ️"
		}
	case doctor.StatusSkipped:
		return "⊘"
	default:
		return "?"
	}
}

// countResults counts errors, warnings, and passed checks
func countResults(results []doctor.CheckResult) (errors, warnings, passed int) {
	for _, result := range results {
		switch {
		case result.IsPassed():
			passed++
		case result.IsError():
			errors++
		case result.IsWarning():
			warnings++
		}
	}

	return errors, warnings, passed
}
