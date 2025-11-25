package provider_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smykla-labs/klaudiush/internal/config"
	"github.com/smykla-labs/klaudiush/internal/config/provider"
	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

// ConfigCombination represents a test scenario for config loading.
type ConfigCombination struct {
	Name          string
	HasGlobal     bool
	HasProject    bool
	HasEnvVars    bool
	HasFlags      bool
	GlobalConfig  *pkgconfig.Config
	ProjectConfig *pkgconfig.Config
	EnvVars       map[string]string
	Flags         map[string]any
}

// configCombinations returns all possible config combinations for testing.
func configCombinations() []ConfigCombination {
	enabled := true
	disabled := false

	globalCfg := &pkgconfig.Config{
		Global: &pkgconfig.GlobalConfig{
			UseSDKGit:      &enabled,
			DefaultTimeout: pkgconfig.Duration(15 * time.Second),
		},
		Validators: &pkgconfig.ValidatorsConfig{
			Git: &pkgconfig.GitConfig{
				Commit: &pkgconfig.CommitValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
				},
			},
		},
	}

	projectCfg := &pkgconfig.Config{
		Validators: &pkgconfig.ValidatorsConfig{
			Git: &pkgconfig.GitConfig{
				Commit: &pkgconfig.CommitValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &disabled,
						Severity: pkgconfig.SeverityWarning,
					},
				},
			},
			File: &pkgconfig.FileConfig{
				Markdown: &pkgconfig.MarkdownValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled: &enabled,
					},
				},
			},
		},
	}

	envVars := map[string]string{
		"KLAUDIUSH_USE_SDK_GIT":                   "false",
		"KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED": "true",
	}

	flags := map[string]any{
		"disable": "markdown,push",
		"timeout": "20s",
	}

	return []ConfigCombination{
		// Single sources
		{
			Name:       "defaults_only",
			HasGlobal:  false,
			HasProject: false,
			HasEnvVars: false,
			HasFlags:   false,
		},
		{
			Name:         "global_only",
			HasGlobal:    true,
			HasProject:   false,
			HasEnvVars:   false,
			HasFlags:     false,
			GlobalConfig: globalCfg,
		},
		{
			Name:          "project_only",
			HasGlobal:     false,
			HasProject:    true,
			HasEnvVars:    false,
			HasFlags:      false,
			ProjectConfig: projectCfg,
		},
		{
			Name:       "env_only",
			HasGlobal:  false,
			HasProject: false,
			HasEnvVars: true,
			HasFlags:   false,
			EnvVars:    envVars,
		},
		{
			Name:       "flags_only",
			HasGlobal:  false,
			HasProject: false,
			HasEnvVars: false,
			HasFlags:   true,
			Flags:      flags,
		},

		// Two sources
		{
			Name:          "global+project",
			HasGlobal:     true,
			HasProject:    true,
			HasEnvVars:    false,
			HasFlags:      false,
			GlobalConfig:  globalCfg,
			ProjectConfig: projectCfg,
		},
		{
			Name:         "global+env",
			HasGlobal:    true,
			HasProject:   false,
			HasEnvVars:   true,
			HasFlags:     false,
			GlobalConfig: globalCfg,
			EnvVars:      envVars,
		},
		{
			Name:         "global+flags",
			HasGlobal:    true,
			HasProject:   false,
			HasEnvVars:   false,
			HasFlags:     true,
			GlobalConfig: globalCfg,
			Flags:        flags,
		},
		{
			Name:          "project+env",
			HasGlobal:     false,
			HasProject:    true,
			HasEnvVars:    true,
			HasFlags:      false,
			ProjectConfig: projectCfg,
			EnvVars:       envVars,
		},
		{
			Name:          "project+flags",
			HasGlobal:     false,
			HasProject:    true,
			HasEnvVars:    false,
			HasFlags:      true,
			ProjectConfig: projectCfg,
			Flags:         flags,
		},
		{
			Name:       "env+flags",
			HasGlobal:  false,
			HasProject: false,
			HasEnvVars: true,
			HasFlags:   true,
			EnvVars:    envVars,
			Flags:      flags,
		},

		// Three sources
		{
			Name:          "global+project+env",
			HasGlobal:     true,
			HasProject:    true,
			HasEnvVars:    true,
			HasFlags:      false,
			GlobalConfig:  globalCfg,
			ProjectConfig: projectCfg,
			EnvVars:       envVars,
		},
		{
			Name:          "global+project+flags",
			HasGlobal:     true,
			HasProject:    true,
			HasEnvVars:    false,
			HasFlags:      true,
			GlobalConfig:  globalCfg,
			ProjectConfig: projectCfg,
			Flags:         flags,
		},
		{
			Name:         "global+env+flags",
			HasGlobal:    true,
			HasProject:   false,
			HasEnvVars:   true,
			HasFlags:     true,
			GlobalConfig: globalCfg,
			EnvVars:      envVars,
			Flags:        flags,
		},
		{
			Name:          "project+env+flags",
			HasGlobal:     false,
			HasProject:    true,
			HasEnvVars:    true,
			HasFlags:      true,
			ProjectConfig: projectCfg,
			EnvVars:       envVars,
			Flags:         flags,
		},

		// All sources
		{
			Name:          "all_sources",
			HasGlobal:     true,
			HasProject:    true,
			HasEnvVars:    true,
			HasFlags:      true,
			GlobalConfig:  globalCfg,
			ProjectConfig: projectCfg,
			EnvVars:       envVars,
			Flags:         flags,
		},
	}
}

// BenchmarkConfigCombinations benchmarks all config source combinations.
func BenchmarkConfigCombinations(b *testing.B) {
	for _, combo := range configCombinations() {
		b.Run(combo.Name, func(b *testing.B) {
			// Setup temp directory for config files
			tmpDir := b.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			projectDir := filepath.Join(tmpDir, "project")

			if err := os.MkdirAll(filepath.Join(homeDir, ".klaudiush"), 0o755); err != nil {
				b.Fatal(err)
			}

			if err := os.MkdirAll(filepath.Join(projectDir, ".klaudiush"), 0o755); err != nil {
				b.Fatal(err)
			}

			// Write config files if needed
			if combo.HasGlobal && combo.GlobalConfig != nil {
				writeConfigFile(
					b,
					filepath.Join(homeDir, ".klaudiush", "config.toml"),
					combo.GlobalConfig,
				)
			}

			if combo.HasProject && combo.ProjectConfig != nil {
				writeConfigFile(
					b,
					filepath.Join(projectDir, ".klaudiush", "config.toml"),
					combo.ProjectConfig,
				)
			}

			// Set/unset env vars
			originalEnv := captureEnv()

			clearKlaudiushEnv()

			if combo.HasEnvVars {
				for k, v := range combo.EnvVars {
					os.Setenv(k, v)
				}
			}

			defer restoreEnv(originalEnv)

			// Build sources
			var sources []provider.Source

			if combo.HasFlags {
				sources = append(sources, provider.NewFlagSource(combo.Flags))
			}

			sources = append(sources, provider.NewEnvSource())

			loader := config.NewLoaderWithDirs(homeDir, projectDir)
			sources = append(sources, provider.NewProjectFileSource(loader))
			sources = append(sources, provider.NewGlobalFileSource(loader))

			p := provider.NewProvider(sources...)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				p.Reload()

				_, err := p.Load()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMergeOperations benchmarks the config merge operation.
func BenchmarkMergeOperations(b *testing.B) {
	enabled := true
	disabled := false

	configs := []*pkgconfig.Config{
		config.DefaultConfig(),
		{
			Global: &pkgconfig.GlobalConfig{
				UseSDKGit: &enabled,
			},
		},
		{
			Validators: &pkgconfig.ValidatorsConfig{
				Git: &pkgconfig.GitConfig{
					Commit: &pkgconfig.CommitValidatorConfig{
						ValidatorConfig: pkgconfig.ValidatorConfig{
							Enabled: &disabled,
						},
					},
				},
			},
		},
		{
			Validators: &pkgconfig.ValidatorsConfig{
				File: &pkgconfig.FileConfig{
					Markdown: &pkgconfig.MarkdownValidatorConfig{
						ValidatorConfig: pkgconfig.ValidatorConfig{
							Enabled: &enabled,
						},
					},
				},
			},
		},
	}

	b.Run("merge_2_configs", func(b *testing.B) {
		merger := config.NewMerger()

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_ = merger.Merge(configs[0], configs[1])
		}
	})

	b.Run("merge_3_configs", func(b *testing.B) {
		merger := config.NewMerger()

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_ = merger.Merge(configs[0], configs[1], configs[2])
		}
	})

	b.Run("merge_4_configs", func(b *testing.B) {
		merger := config.NewMerger()

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_ = merger.Merge(configs...)
		}
	})
}

// BenchmarkEnvSourceParsing benchmarks env var parsing.
func BenchmarkEnvSourceParsing(b *testing.B) {
	scenarios := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name:    "no_env_vars",
			envVars: map[string]string{},
		},
		{
			name: "single_env_var",
			envVars: map[string]string{
				"KLAUDIUSH_USE_SDK_GIT": "true",
			},
		},
		{
			name: "few_env_vars",
			envVars: map[string]string{
				"KLAUDIUSH_USE_SDK_GIT":                   "true",
				"KLAUDIUSH_DEFAULT_TIMEOUT":               "15s",
				"KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED": "false",
			},
		},
		{
			name: "many_env_vars",
			envVars: map[string]string{
				"KLAUDIUSH_USE_SDK_GIT":                          "true",
				"KLAUDIUSH_DEFAULT_TIMEOUT":                      "15s",
				"KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED":        "false",
				"KLAUDIUSH_VALIDATORS_GIT_COMMIT_SEVERITY":       "warning",
				"KLAUDIUSH_VALIDATORS_GIT_PUSH_ENABLED":          "true",
				"KLAUDIUSH_VALIDATORS_GIT_ADD_ENABLED":           "false",
				"KLAUDIUSH_VALIDATORS_FILE_MARKDOWN_ENABLED":     "true",
				"KLAUDIUSH_VALIDATORS_FILE_SHELLSCRIPT_ENABLED":  "false",
				"KLAUDIUSH_VALIDATORS_NOTIFICATION_BELL_ENABLED": "true",
			},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			originalEnv := captureEnv()

			clearKlaudiushEnv()

			for k, v := range scenario.envVars {
				os.Setenv(k, v)
			}

			defer restoreEnv(originalEnv)

			source := provider.NewEnvSource()

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_, _ = source.Load()
			}
		})
	}
}

// BenchmarkFlagSourceParsing benchmarks CLI flag parsing.
func BenchmarkFlagSourceParsing(b *testing.B) {
	scenarios := []struct {
		name  string
		flags map[string]any
	}{
		{
			name:  "no_flags",
			flags: map[string]any{},
		},
		{
			name: "single_flag",
			flags: map[string]any{
				"use-sdk-git": true,
			},
		},
		{
			name: "few_flags",
			flags: map[string]any{
				"use-sdk-git": true,
				"timeout":     "15s",
				"disable":     "commit",
			},
		},
		{
			name: "many_flags",
			flags: map[string]any{
				"use-sdk-git": true,
				"timeout":     "15s",
				"disable":     "commit,push,add,markdown,shellscript,terraform",
			},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			source := provider.NewFlagSource(scenario.flags)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_, _ = source.Load()
			}
		})
	}
}

// BenchmarkCacheHitVsMiss benchmarks cache hit vs miss performance.
func BenchmarkCacheHitVsMiss(b *testing.B) {
	p := provider.NewProvider()

	// Warm up cache
	if _, err := p.Load(); err != nil {
		b.Fatal(err)
	}

	b.Run("cache_hit", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_, err := p.Load()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("cache_miss", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			p.Reload()

			_, err := p.Load()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkFileLoading benchmarks config file I/O.
func BenchmarkFileLoading(b *testing.B) {
	scenarios := []struct {
		name   string
		config *pkgconfig.Config
	}{
		{
			name:   "minimal_config",
			config: &pkgconfig.Config{},
		},
		{
			name:   "default_config",
			config: config.DefaultConfig(),
		},
		{
			name:   "full_config",
			config: createFullConfig(),
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			tmpDir := b.TempDir()
			homeDir := filepath.Join(tmpDir, "home")

			if err := os.MkdirAll(filepath.Join(homeDir, ".klaudiush"), 0o755); err != nil {
				b.Fatal(err)
			}

			configPath := filepath.Join(homeDir, ".klaudiush", "config.toml")
			writeConfigFile(b, configPath, scenario.config)

			loader := config.NewLoaderWithDirs(homeDir, tmpDir)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_, err := loader.LoadGlobal()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkValidatorEnableDisable benchmarks enabling/disabling validators.
func BenchmarkValidatorEnableDisable(b *testing.B) {
	scenarios := []struct {
		name    string
		disable string
	}{
		{name: "none_disabled", disable: ""},
		{name: "one_disabled", disable: "commit"},
		{name: "few_disabled", disable: "commit,push,markdown"},
		{name: "many_disabled", disable: "commit,push,add,markdown,shellscript,terraform"},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			var flags map[string]any
			if scenario.disable != "" {
				flags = map[string]any{"disable": scenario.disable}
			}

			p := provider.NewProvider(provider.NewFlagSource(flags))

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				p.Reload()

				_, err := p.Load()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkValidation benchmarks config validation.
func BenchmarkValidation(b *testing.B) {
	scenarios := []struct {
		name   string
		config *pkgconfig.Config
	}{
		{name: "empty_config", config: &pkgconfig.Config{}},
		{name: "default_config", config: config.DefaultConfig()},
		{name: "full_config", config: createFullConfig()},
	}

	validator := config.NewValidator()

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_ = validator.Validate(scenario.config)
			}
		})
	}
}

// BenchmarkDefaultConfigCreation benchmarks creating default configs.
func BenchmarkDefaultConfigCreation(b *testing.B) {
	b.Run("full_default", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_ = config.DefaultConfig()
		}
	})

	b.Run("global_default", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_ = config.DefaultGlobalConfig()
		}
	})

	b.Run("validators_default", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_ = config.DefaultValidatorsConfig()
		}
	})
}

// Helper functions

func writeConfigFile(tb testing.TB, path string, cfg *pkgconfig.Config) {
	tb.Helper()

	content := configToTOML(cfg)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		tb.Fatal(err)
	}
}

func configToTOML(cfg *pkgconfig.Config) string {
	var content string

	if cfg.Global != nil {
		content += "[global]\n"

		if cfg.Global.UseSDKGit != nil {
			content += fmt.Sprintf("use_sdk_git = %t\n", *cfg.Global.UseSDKGit)
		}

		if cfg.Global.DefaultTimeout != 0 {
			content += fmt.Sprintf(
				"default_timeout = %q\n",
				time.Duration(cfg.Global.DefaultTimeout).String(),
			)
		}

		content += "\n"
	}

	if cfg.Validators != nil && cfg.Validators.Git != nil {
		if cfg.Validators.Git.Commit != nil {
			content += "[validators.git.commit]\n"

			if cfg.Validators.Git.Commit.Enabled != nil {
				content += fmt.Sprintf("enabled = %t\n", *cfg.Validators.Git.Commit.Enabled)
			}

			if cfg.Validators.Git.Commit.Severity != pkgconfig.SeverityUnknown {
				content += fmt.Sprintf(
					"severity = %q\n",
					cfg.Validators.Git.Commit.Severity.String(),
				)
			}

			content += "\n"
		}

		if cfg.Validators.Git.Push != nil {
			content += "[validators.git.push]\n"

			if cfg.Validators.Git.Push.Enabled != nil {
				content += fmt.Sprintf("enabled = %t\n", *cfg.Validators.Git.Push.Enabled)
			}

			content += "\n"
		}
	}

	if cfg.Validators != nil && cfg.Validators.File != nil {
		if cfg.Validators.File.Markdown != nil {
			content += "[validators.file.markdown]\n"

			if cfg.Validators.File.Markdown.Enabled != nil {
				content += fmt.Sprintf("enabled = %t\n", *cfg.Validators.File.Markdown.Enabled)
			}

			content += "\n"
		}
	}

	return content
}

func captureEnv() map[string]string {
	env := make(map[string]string)

	for _, e := range os.Environ() {
		for i := range e {
			if e[i] == '=' {
				env[e[:i]] = e[i+1:]

				break
			}
		}
	}

	return env
}

func clearKlaudiushEnv() {
	for _, e := range os.Environ() {
		for i := range e {
			if e[i] == '=' {
				key := e[:i]
				if len(key) >= 10 && key[:10] == "KLAUDIUSH_" {
					os.Unsetenv(key)
				}

				break
			}
		}
	}
}

func restoreEnv(original map[string]string) {
	clearKlaudiushEnv()

	for k, v := range original {
		if len(k) >= 10 && k[:10] == "KLAUDIUSH_" {
			os.Setenv(k, v)
		}
	}
}

func createFullConfig() *pkgconfig.Config {
	enabled := true
	disabled := false
	titleMax := 50
	bodyMax := 72
	contextLines := 2
	tolerance := 5
	requireScope := true
	blockInfra := true
	blockPR := true
	blockAI := true
	requireTracking := true
	headingSpacing := true
	codeBlockFmt := true
	listFmt := true
	useMarkdownlint := true
	useShellcheck := true
	checkFormat := true
	useTflint := true
	enforceDigest := true
	requireVersion := true
	checkLatest := true
	useActionlint := true
	conventionalCommits := true
	requireType := true
	allowUppercase := false
	requireBody := true
	requireChangelog := false
	checkCILabels := true
	titleConventional := true

	return &pkgconfig.Config{
		Global: &pkgconfig.GlobalConfig{
			UseSDKGit:      &enabled,
			DefaultTimeout: pkgconfig.Duration(10 * time.Second),
		},
		Validators: &pkgconfig.ValidatorsConfig{
			Git: &pkgconfig.GitConfig{
				Commit: &pkgconfig.CommitValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					RequiredFlags:    []string{"-s", "-S"},
					CheckStagingArea: &enabled,
					Message: &pkgconfig.CommitMessageConfig{
						Enabled:               &enabled,
						TitleMaxLength:        &titleMax,
						BodyMaxLineLength:     &bodyMax,
						BodyLineTolerance:     &tolerance,
						ConventionalCommits:   &conventionalCommits,
						RequireScope:          &requireScope,
						BlockInfraScopeMisuse: &blockInfra,
						BlockPRReferences:     &blockPR,
						BlockAIAttribution:    &blockAI,
						ValidTypes: []string{
							"feat",
							"fix",
							"docs",
							"style",
							"refactor",
							"perf",
							"test",
							"build",
							"ci",
							"chore",
							"revert",
						},
					},
				},
				Push: &pkgconfig.PushValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					RequireTracking: &requireTracking,
				},
				Add: &pkgconfig.AddValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					BlockedPatterns: []string{"tmp/*", "*.log"},
				},
				PR: &pkgconfig.PRValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					TitleMaxLength:           &titleMax,
					TitleConventionalCommits: &titleConventional,
					RequireChangelog:         &requireChangelog,
					CheckCILabels:            &checkCILabels,
					RequireBody:              &requireBody,
					ValidTypes: []string{
						"feat",
						"fix",
						"docs",
						"style",
						"refactor",
						"perf",
						"test",
						"build",
						"ci",
						"chore",
						"revert",
					},
					MarkdownDisabledRules: []string{"MD013", "MD034", "MD041"},
				},
				Branch: &pkgconfig.BranchValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					ProtectedBranches: []string{"main", "master"},
					RequireType:       &requireType,
					AllowUppercase:    &allowUppercase,
					ValidTypes: []string{
						"feat",
						"fix",
						"docs",
						"style",
						"refactor",
						"perf",
						"test",
						"build",
						"ci",
						"chore",
					},
				},
				NoVerify: &pkgconfig.NoVerifyValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
				},
			},
			File: &pkgconfig.FileConfig{
				Markdown: &pkgconfig.MarkdownValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					Timeout:             pkgconfig.Duration(10 * time.Second),
					ContextLines:        &contextLines,
					HeadingSpacing:      &headingSpacing,
					CodeBlockFormatting: &codeBlockFmt,
					ListFormatting:      &listFmt,
					UseMarkdownlint:     &useMarkdownlint,
				},
				ShellScript: &pkgconfig.ShellScriptValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					Timeout:            pkgconfig.Duration(10 * time.Second),
					ContextLines:       &contextLines,
					UseShellcheck:      &useShellcheck,
					ShellcheckSeverity: "warning",
				},
				Terraform: &pkgconfig.TerraformValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &disabled,
						Severity: pkgconfig.SeverityError,
					},
					Timeout:        pkgconfig.Duration(30 * time.Second),
					ContextLines:   &contextLines,
					ToolPreference: "auto",
					CheckFormat:    &checkFormat,
					UseTflint:      &useTflint,
				},
				Workflow: &pkgconfig.WorkflowValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
					Timeout:               pkgconfig.Duration(30 * time.Second),
					GHAPITimeout:          pkgconfig.Duration(5 * time.Second),
					EnforceDigestPinning:  &enforceDigest,
					RequireVersionComment: &requireVersion,
					CheckLatestVersion:    &checkLatest,
					UseActionlint:         &useActionlint,
				},
			},
			Notification: &pkgconfig.NotificationConfig{
				Bell: &pkgconfig.BellValidatorConfig{
					ValidatorConfig: pkgconfig.ValidatorConfig{
						Enabled:  &enabled,
						Severity: pkgconfig.SeverityError,
					},
				},
			},
		},
	}
}
