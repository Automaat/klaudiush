// Package hook provides core types for Claude Code hook context.
package hook

import "encoding/json"

// EventType represents the type of hook event.
type EventType string

const (
	// PreToolUse is triggered before a tool is executed.
	PreToolUse EventType = "PreToolUse"

	// PostToolUse is triggered after a tool is executed.
	PostToolUse EventType = "PostToolUse"

	// Notification is triggered for user notifications.
	Notification EventType = "Notification"
)

// ToolType represents the type of tool being used.
type ToolType string

const (
	// Bash represents the Bash tool for executing shell commands.
	Bash ToolType = "Bash"

	// Write represents the Write tool for creating files.
	Write ToolType = "Write"

	// Edit represents the Edit tool for modifying files.
	Edit ToolType = "Edit"

	// MultiEdit represents the MultiEdit tool for modifying multiple files.
	MultiEdit ToolType = "MultiEdit"

	// Grep represents the Grep tool for searching files.
	Grep ToolType = "Grep"

	// Read represents the Read tool for reading files.
	Read ToolType = "Read"

	// Glob represents the Glob tool for finding files by pattern.
	Glob ToolType = "Glob"
)

// ToolInput contains the raw tool input data.
type ToolInput struct {
	// Command is the shell command for Bash tool.
	Command string `json:"command,omitempty"`

	// FilePath is the file path for file operations.
	FilePath string `json:"file_path,omitempty"`

	// Path is an alternative field for file path.
	Path string `json:"path,omitempty"`

	// Content is the file content for Write tool.
	Content string `json:"content,omitempty"`

	// OldString is the string to replace for Edit tool.
	OldString string `json:"old_string,omitempty"`

	// NewString is the replacement string for Edit tool.
	NewString string `json:"new_string,omitempty"`

	// Pattern is the search pattern for Grep/Glob tools.
	Pattern string `json:"pattern,omitempty"`

	// Additional fields stored as raw JSON.
	Additional map[string]json.RawMessage `json:"-"`
}

// Context represents the complete hook invocation context.
type Context struct {
	// EventType is the type of hook event (PreToolUse, PostToolUse, Notification).
	EventType EventType

	// ToolName is the name of the tool being invoked.
	ToolName ToolType

	// ToolInput contains the tool-specific input parameters.
	ToolInput ToolInput

	// NotificationType is the type of notification (for Notification events).
	NotificationType string

	// RawJSON contains the original JSON input for advanced parsing.
	RawJSON string
}

// GetCommand returns the command from ToolInput.
func (c *Context) GetCommand() string {
	return c.ToolInput.Command
}

// GetFilePath returns the file path from ToolInput, preferring FilePath over Path.
func (c *Context) GetFilePath() string {
	if c.ToolInput.FilePath != "" {
		return c.ToolInput.FilePath
	}

	return c.ToolInput.Path
}

// GetContent returns the file content from ToolInput.
func (c *Context) GetContent() string {
	return c.ToolInput.Content
}

// IsBashTool returns true if the tool is Bash.
func (c *Context) IsBashTool() bool {
	return c.ToolName == Bash
}

// IsFileTool returns true if the tool is a file operation (Write, Edit, MultiEdit).
func (c *Context) IsFileTool() bool {
	return c.ToolName == Write || c.ToolName == Edit || c.ToolName == MultiEdit
}
