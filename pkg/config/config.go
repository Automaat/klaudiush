// Package config provides configuration schema types for klaudiush validators.
package config

// Config represents the root configuration for klaudiush.
type Config struct {
	// Validators groups all validator configurations.
	Validators *ValidatorsConfig `json:"validators,omitempty" toml:"validators"`

	// Global settings that apply across all validators.
	Global *GlobalConfig `json:"global,omitempty" toml:"global"`
}

// ValidatorsConfig groups all validator configurations by category.
type ValidatorsConfig struct {
	// Git validator configurations.
	Git *GitConfig `json:"git,omitempty" toml:"git"`

	// File validator configurations.
	File *FileConfig `json:"file,omitempty" toml:"file"`

	// Notification validator configurations.
	Notification *NotificationConfig `json:"notification,omitempty" toml:"notification"`
}

// GlobalConfig contains global settings that apply to all validators.
type GlobalConfig struct {
	// UseSDKGit controls whether to use the go-git SDK or CLI for git operations.
	// Default: true (use SDK for better performance)
	UseSDKGit *bool `json:"use_sdk_git,omitempty" toml:"use_sdk_git"`

	// DefaultTimeout is the default timeout for all operations that support timeouts.
	// Individual validator timeouts override this value.
	// Default: "10s"
	DefaultTimeout Duration `json:"default_timeout,omitempty" toml:"default_timeout"`
}

// GetValidators returns the validators config, creating it if it doesn't exist.
func (c *Config) GetValidators() *ValidatorsConfig {
	if c.Validators == nil {
		c.Validators = &ValidatorsConfig{}
	}

	return c.Validators
}

// GetGlobal returns the global config, creating it if it doesn't exist.
func (c *Config) GetGlobal() *GlobalConfig {
	if c.Global == nil {
		c.Global = &GlobalConfig{}
	}

	return c.Global
}

// GetGit returns the git validators config, creating it if it doesn't exist.
func (v *ValidatorsConfig) GetGit() *GitConfig {
	if v.Git == nil {
		v.Git = &GitConfig{}
	}

	return v.Git
}

// GetFile returns the file validators config, creating it if it doesn't exist.
func (v *ValidatorsConfig) GetFile() *FileConfig {
	if v.File == nil {
		v.File = &FileConfig{}
	}

	return v.File
}

// GetNotification returns the notification validators config, creating it if it doesn't exist.
func (v *ValidatorsConfig) GetNotification() *NotificationConfig {
	if v.Notification == nil {
		v.Notification = &NotificationConfig{}
	}

	return v.Notification
}
