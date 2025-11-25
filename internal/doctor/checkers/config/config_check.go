// Package config provides checkers for configuration file validation.
package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/smykla-labs/klaudiush/internal/config"
	"github.com/smykla-labs/klaudiush/internal/doctor"
)

// GlobalChecker checks the validity of the global configuration
type GlobalChecker struct {
	loader *config.KoanfLoader
}

// NewGlobalChecker creates a new global config checker
func NewGlobalChecker() *GlobalChecker {
	loader, _ := config.NewKoanfLoader()

	return &GlobalChecker{
		loader: loader,
	}
}

// Name returns the name of the check
func (*GlobalChecker) Name() string {
	return "Global config"
}

// Category returns the category of the check
func (*GlobalChecker) Category() doctor.Category {
	return doctor.CategoryConfig
}

// Check performs the global config validity check
func (c *GlobalChecker) Check(_ context.Context) doctor.CheckResult {
	if !c.loader.HasGlobalConfig() {
		return doctor.FailWarning("Global config", "Not found (optional)").
			WithDetails(
				"Expected at: "+c.loader.GlobalConfigPath(),
				"Create with: klaudiush init --global",
			).
			WithFixID("create_global_config")
	}

	// Try loading config to validate it
	cfg, err := c.loader.Load(nil)
	if err != nil {
		if errors.Is(err, config.ErrInvalidTOML) {
			return doctor.FailError("Global config", "Invalid TOML syntax").
				WithDetails(
					"File: "+c.loader.GlobalConfigPath(),
					fmt.Sprintf("Error: %v", err),
				)
		}

		if errors.Is(err, config.ErrInvalidPermissions) {
			return doctor.FailError("Global config", "Insecure file permissions").
				WithDetails(
					"File: "+c.loader.GlobalConfigPath(),
					"Config file should not be world-writable",
					"Fix with: chmod 600 <config-file>",
				).
				WithFixID("fix_config_permissions")
		}

		return doctor.FailError("Global config", fmt.Sprintf("Failed to load: %v", err))
	}

	// Validate config semantics
	validator := config.NewValidator()
	if err := validator.Validate(cfg); err != nil {
		return doctor.FailError("Global config", "Validation failed").
			WithDetails(
				"File: "+c.loader.GlobalConfigPath(),
				fmt.Sprintf("Error: %v", err),
			)
	}

	return doctor.Pass("Global config", "Loaded and validated")
}

// ProjectChecker checks the validity of the project configuration
type ProjectChecker struct {
	loader *config.KoanfLoader
}

// NewProjectChecker creates a new project config checker
func NewProjectChecker() *ProjectChecker {
	loader, _ := config.NewKoanfLoader()

	return &ProjectChecker{
		loader: loader,
	}
}

// Name returns the name of the check
func (*ProjectChecker) Name() string {
	return "Project config"
}

// Category returns the category of the check
func (*ProjectChecker) Category() doctor.Category {
	return doctor.CategoryConfig
}

// Check performs the project config validity check
func (c *ProjectChecker) Check(_ context.Context) doctor.CheckResult {
	if !c.loader.HasProjectConfig() {
		// Project config not found is just informational since global config is the primary
		return doctor.Skip("Project config", "Not found (using global config)")
	}

	cfg, err := c.loader.Load(nil)
	if err != nil {
		if errors.Is(err, config.ErrInvalidTOML) {
			return doctor.FailError("Project config", "Invalid TOML syntax").
				WithDetails(fmt.Sprintf("Error: %v", err))
		}

		if errors.Is(err, config.ErrInvalidPermissions) {
			return doctor.FailError("Project config", "Insecure file permissions").
				WithDetails(
					"Config file should not be world-writable",
					"Fix with: chmod 600 <config-file>",
				).
				WithFixID("fix_config_permissions")
		}

		return doctor.FailError("Project config", fmt.Sprintf("Failed to load: %v", err))
	}

	// Validate config semantics
	validator := config.NewValidator()
	if err := validator.Validate(cfg); err != nil {
		return doctor.FailError("Project config", "Validation failed").
			WithDetails(fmt.Sprintf("Error: %v", err))
	}

	return doctor.Pass("Project config", "Loaded and validated")
}

// PermissionsChecker checks if config files have secure permissions
type PermissionsChecker struct {
	loader *config.KoanfLoader
}

// NewPermissionsChecker creates a new permissions checker
func NewPermissionsChecker() *PermissionsChecker {
	loader, _ := config.NewKoanfLoader()

	return &PermissionsChecker{
		loader: loader,
	}
}

// Name returns the name of the check
func (*PermissionsChecker) Name() string {
	return "Config permissions"
}

// Category returns the category of the check
func (*PermissionsChecker) Category() doctor.Category {
	return doctor.CategoryConfig
}

// Check performs the permissions check
func (c *PermissionsChecker) Check(_ context.Context) doctor.CheckResult {
	// Check both global and project configs
	hasGlobal := c.loader.HasGlobalConfig()
	hasProject := c.loader.HasProjectConfig()

	if !hasGlobal && !hasProject {
		return doctor.Skip("Config permissions", "No config files found")
	}

	// Try loading - if they have permission issues, they'll fail
	_, err := c.loader.Load(nil)

	// Check for permission errors
	if err != nil && errors.Is(err, config.ErrInvalidPermissions) {
		return doctor.FailError("Config permissions", "Insecure file permissions detected").
			WithDetails(
				fmt.Sprintf("Error: %v", err),
				"Fix with: chmod 600 <config-file>",
			).
			WithFixID("fix_config_permissions")
	}

	return doctor.Pass("Config permissions", "Files are secured")
}
