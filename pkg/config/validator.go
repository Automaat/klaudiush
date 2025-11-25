// Package config provides configuration schema types for klaudiush validators.
package config

// ValidatorConfig represents the base configuration for all validators.
type ValidatorConfig struct {
	// Enabled controls whether the validator is active.
	// When false, the validator is completely skipped.
	// Default: true
	Enabled *bool `json:"enabled,omitempty" toml:"enabled"`

	// Severity determines whether validation failures block the operation.
	// "error" blocks the operation (default)
	// "warning" only warns without blocking
	Severity Severity `json:"severity,omitempty" toml:"severity"`
}

// IsEnabled returns true if the validator is enabled.
// Returns true if Enabled is nil (default behavior).
func (c *ValidatorConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}

	return *c.Enabled
}

// GetSeverity returns the severity level, defaulting to Error if not set.
func (c *ValidatorConfig) GetSeverity() Severity {
	if c.Severity == SeverityUnknown {
		return SeverityError
	}

	return c.Severity
}
