package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/smykla-labs/claude-hooks/internal/linters"
	"github.com/smykla-labs/claude-hooks/internal/validator"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

const (
	shellCheckTimeout = 10 * time.Second
)

// ShellScriptValidator validates shell scripts using shellcheck.
type ShellScriptValidator struct {
	validator.BaseValidator
	checker linters.ShellChecker
}

// NewShellScriptValidator creates a new ShellScriptValidator.
func NewShellScriptValidator(
	log logger.Logger,
	checker linters.ShellChecker,
) *ShellScriptValidator {
	return &ShellScriptValidator{
		BaseValidator: *validator.NewBaseValidator("validate-shellscript", log),
		checker:       checker,
	}
}

// Validate validates shell scripts using shellcheck.
func (v *ShellScriptValidator) Validate(ctx *hook.Context) *validator.Result {
	log := v.Logger()
	log.Debug("validating shell script")

	// Get the file path
	filePath := ctx.GetFilePath()
	if filePath == "" {
		log.Debug("no file path provided")
		return validator.Pass()
	}

	// Get content from context or read from file
	content := ctx.ToolInput.Content
	if content == "" {
		// Check if file exists
		if _, err := os.Stat(filePath); err != nil {
			log.Debug("file does not exist, skipping", "file", filePath)
			return validator.Pass()
		}

		// Read file content
		data, err := os.ReadFile(filePath) //nolint:gosec // filePath is from Claude Code context
		if err != nil {
			log.Debug("failed to read file", "file", filePath, "error", err)
			return validator.Pass()
		}

		content = string(data)
	}

	// Skip Fish scripts
	if v.isFishScript(filePath, content) {
		log.Debug("skipping Fish script", "file", filePath)
		return validator.Pass()
	}

	// Run shellcheck using the linter
	lintCtx, cancel := context.WithTimeout(context.Background(), shellCheckTimeout)
	defer cancel()

	result := v.checker.Check(lintCtx, content)
	if result.Success {
		log.Debug("shellcheck passed")
		return validator.Pass()
	}

	log.Debug("shellcheck failed", "output", result.RawOut)

	return validator.Fail(v.formatShellCheckOutput(result.RawOut))
}

// isFishScript checks if the script is a Fish shell script.
func (v *ShellScriptValidator) isFishScript(filePath, content string) bool {
	// Check file extension
	if filepath.Ext(filePath) == ".fish" {
		return true
	}

	// Check shebang
	if strings.HasPrefix(content, "#!/usr/bin/env fish") ||
		strings.HasPrefix(content, "#!/usr/bin/fish") ||
		strings.HasPrefix(content, "#!/bin/fish") {
		return true
	}

	return false
}

// formatShellCheckOutput formats shellcheck output for display.
func (v *ShellScriptValidator) formatShellCheckOutput(output string) string {
	// Clean up the output - remove empty lines
	lines := strings.Split(output, "\n")

	var cleanLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return "Shellcheck validation failed\n\n" + strings.Join(
		cleanLines,
		"\n",
	) + "\n\nFix these issues before committing."
}
