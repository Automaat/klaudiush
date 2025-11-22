package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/smykla-labs/claude-hooks/internal/validator"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
	"github.com/smykla-labs/claude-hooks/pkg/parser"
)

// BranchValidator validates git branch names.
type BranchValidator struct {
	validator.BaseValidator
}

// NewBranchValidator creates a new BranchValidator.
func NewBranchValidator(log logger.Logger) *BranchValidator {
	return &BranchValidator{
		BaseValidator: *validator.NewBaseValidator("validate-branch-name", log),
	}
}

const (
	// minBranchParts is the minimum number of parts in a valid branch name.
	minBranchParts = 2
)

var (
	// Valid branch name pattern: type/description (e.g., feat/add-feature, fix/bug-123).
	branchNamePattern = regexp.MustCompile(`^[a-z]+/[a-z0-9-]+$`)

	// Protected branches that should skip validation.
	protectedBranches = map[string]bool{
		"main":   true,
		"master": true,
	}

	// Valid branch types.
	validBranchTypes = map[string]bool{
		"feat":     true,
		"fix":      true,
		"docs":     true,
		"style":    true,
		"refactor": true,
		"test":     true,
		"chore":    true,
		"ci":       true,
		"build":    true,
		"perf":     true,
	}
)

// Validate validates git branch names.
func (v *BranchValidator) Validate(ctx *hook.Context) *validator.Result {
	log := v.Logger()
	log.Debug("validating git branch command")

	bashParser := parser.NewBashParser()
	parseResult, err := bashParser.Parse(ctx.ToolInput.Command)
	if err != nil {
		log.Error("failed to parse command", "error", err)
		return validator.Warn(fmt.Sprintf("Failed to parse command: %v", err))
	}

	for _, cmd := range parseResult.Commands {
		if cmd.Name != "git" {
			continue
		}

		gitCmd, err := parser.ParseGitCommand(cmd)
		if err != nil {
			v.Logger().Debug("failed to parse git command", "error", err)
			continue
		}

		result := v.validateGitCommand(gitCmd)
		if result != nil && !result.Passed {
			return result
		}
	}

	return validator.Pass()
}

// validateGitCommand validates a git command based on its subcommand.
func (v *BranchValidator) validateGitCommand(gitCmd *parser.GitCommand) *validator.Result {
	switch gitCmd.Subcommand {
	case "checkout":
		return v.validateCheckout(gitCmd)
	case "branch":
		return v.validateBranch(gitCmd)
	default:
		return nil
	}
}

// validateCheckout validates git checkout -b commands.
func (v *BranchValidator) validateCheckout(gitCmd *parser.GitCommand) *validator.Result {
	if !gitCmd.HasFlag("-b") {
		return nil
	}

	branchName, hasExtra := v.extractBranchName(gitCmd)
	if branchName == "" {
		return nil
	}

	if hasExtra {
		return v.createSpaceError()
	}

	return v.validateBranchName(branchName)
}

// validateBranch validates git branch commands.
func (v *BranchValidator) validateBranch(gitCmd *parser.GitCommand) *validator.Result {
	// Skip if it's a delete operation.
	if gitCmd.HasFlag("-d") || gitCmd.HasFlag("-D") || gitCmd.HasFlag("--delete") {
		return nil
	}

	branchName, hasExtra := v.extractBranchName(gitCmd)
	if branchName == "" {
		return nil
	}

	if hasExtra {
		return v.createSpaceError()
	}

	return v.validateBranchName(branchName)
}

// createSpaceError creates an error for branch names with spaces.
func (v *BranchValidator) createSpaceError() *validator.Result {
	var msg strings.Builder
	msg.WriteString("Branch name appears to contain spaces\n\n")
	msg.WriteString("Branch names cannot contain spaces. Use hyphens instead.\n\n")
	msg.WriteString("Example: feat/my-feature not feat/my feature")
	return validator.Fail(msg.String())
}

// extractBranchName extracts the branch name from a git command.
// Returns the branch name and a boolean indicating if there are extra arguments
// that suggest the branch name contains spaces.
func (v *BranchValidator) extractBranchName(gitCmd *parser.GitCommand) (string, bool) {
	switch gitCmd.Subcommand {
	case "checkout":
		return v.extractCheckoutBranchName(gitCmd)
	case "branch":
		return v.extractBranchCommandName(gitCmd)
	default:
		return "", false
	}
}

// extractCheckoutBranchName extracts the branch name from git checkout -b.
func (v *BranchValidator) extractCheckoutBranchName(gitCmd *parser.GitCommand) (string, bool) {
	// For 'git checkout -b <branch>', the branch name is after -b.
	for i, flag := range gitCmd.Flags {
		if flag == "-b" && i+1 < len(gitCmd.Flags) {
			branchName := gitCmd.Flags[i+1]
			// Check if there are extra arguments after the branch name
			hasExtra := len(gitCmd.Args) > 0
			return branchName, hasExtra
		}
	}

	// Try args as well.
	if len(gitCmd.Args) > 0 {
		branchName := gitCmd.Args[0]
		// Check if there are extra arguments
		hasExtra := len(gitCmd.Args) > 1
		return branchName, hasExtra
	}

	return "", false
}

// extractBranchCommandName extracts the branch name from git branch.
func (v *BranchValidator) extractBranchCommandName(gitCmd *parser.GitCommand) (string, bool) {
	// For 'git branch <branch>', the branch name is the first arg.
	if len(gitCmd.Args) > 0 {
		branchName := gitCmd.Args[0]
		// Check if there are extra arguments
		hasExtra := len(gitCmd.Args) > 1
		return branchName, hasExtra
	}

	return "", false
}

// validateBranchName validates the branch name format.
func (v *BranchValidator) validateBranchName(branchName string) *validator.Result {
	// Skip protected branches.
	if protectedBranches[branchName] {
		v.Logger().Debug("skipping protected branch", "branch", branchName)
		return validator.Pass()
	}

	// Check for uppercase characters.
	if branchName != strings.ToLower(branchName) {
		var msg strings.Builder
		msg.WriteString("Branch name must be lowercase\n\n")
		fmt.Fprintf(&msg, "Branch name '%s' contains uppercase characters\n\n", branchName)
		fmt.Fprintf(&msg, "Use: %s", strings.ToLower(branchName))
		return validator.Fail(msg.String())
	}

	// Check format: type/description.
	if !branchNamePattern.MatchString(branchName) {
		var msg strings.Builder
		msg.WriteString("Branch name must follow type/description format\n\n")
		fmt.Fprintf(&msg, "Branch name '%s' doesn't match pattern\n\n", branchName)
		msg.WriteString("Expected format: <type>/<description>\n")
		msg.WriteString("Valid types: feat, fix, docs, style, refactor, test, chore, ci, build, perf\n\n")
		msg.WriteString("Example: feat/add-user-auth or fix/login-bug-123")
		return validator.Fail(msg.String())
	}

	// Extract and validate type.
	parts := strings.SplitN(branchName, "/", minBranchParts)
	if len(parts) != minBranchParts {
		var msg strings.Builder
		msg.WriteString("Branch name must contain type and description\n\n")
		fmt.Fprintf(&msg, "Branch name '%s' is missing type or description\n\n", branchName)
		msg.WriteString("Expected format: <type>/<description>")
		return validator.Fail(msg.String())
	}

	branchType := parts[0]
	if !validBranchTypes[branchType] {
		validTypes := make([]string, 0, len(validBranchTypes))
		for t := range validBranchTypes {
			validTypes = append(validTypes, t)
		}

		var msg strings.Builder
		msg.WriteString("Invalid branch type\n\n")
		fmt.Fprintf(&msg, "Branch type '%s' is not valid\n\n", branchType)
		fmt.Fprintf(&msg, "Valid types: %s", strings.Join(validTypes, ", "))
		return validator.Fail(msg.String())
	}

	return validator.Pass()
}
