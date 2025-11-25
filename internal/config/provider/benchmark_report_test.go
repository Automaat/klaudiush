package provider_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/smykla-labs/klaudiush/internal/config"
	"github.com/smykla-labs/klaudiush/internal/config/provider"
	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

// BenchmarkResult holds timing data for a single benchmark scenario.
type BenchmarkResult struct {
	Name     string
	Duration time.Duration
	Allocs   int64
	Bytes    int64
}

// TestConfigPerformanceReport runs all config loading scenarios and prints a comparison table.
// Run with: go test -v -run TestConfigPerformanceReport ./internal/config/provider/
func TestConfigPerformanceReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance report in short mode")
	}

	iterations := 1000

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("Configuration Loading Performance Report")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("Iterations per scenario: %d\n\n", iterations)

	// Section 1: Config source combinations
	fmt.Println("## Config Source Combinations")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	combos := configCombinations()
	comboResults := make([]BenchmarkResult, 0, len(combos))

	for _, combo := range combos {
		result := benchmarkCombination(t, combo, iterations)
		comboResults = append(comboResults, result)
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Source Combinations", comboResults)

	// Section 2: Merge operations
	fmt.Println("\n## Merge Operations")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	mergeResults := benchmarkMergeOps(iterations)
	for _, result := range mergeResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Merge Operations", mergeResults)

	// Section 3: Environment variable parsing
	fmt.Println("\n## Environment Variable Parsing")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	envResults := benchmarkEnvParsing(t, iterations)
	for _, result := range envResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Env Parsing", envResults)

	// Section 4: Flag parsing
	fmt.Println("\n## CLI Flag Parsing")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	flagResults := benchmarkFlagParsing(iterations)
	for _, result := range flagResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Flag Parsing", flagResults)

	// Section 5: Cache performance
	fmt.Println("\n## Cache Performance")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	cacheResults := benchmarkCachePerf(iterations)
	for _, result := range cacheResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Cache Performance", cacheResults)

	// Section 6: File loading
	fmt.Println("\n## File Loading")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	fileResults := benchmarkFileLoadingReport(t, iterations)
	for _, result := range fileResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("File Loading", fileResults)

	// Section 7: Validator enable/disable
	fmt.Println("\n## Validator Enable/Disable")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	validatorResults := benchmarkValidatorToggle(iterations)
	for _, result := range validatorResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Validator Toggle", validatorResults)

	// Section 8: Validation
	fmt.Println("\n## Config Validation")
	fmt.Println(strings.Repeat("-", 100))
	printTableHeader()

	validationResults := benchmarkValidationReport(iterations)
	for _, result := range validationResults {
		printTableRow(result)
	}

	fmt.Println(strings.Repeat("-", 100))
	printSummary("Validation", validationResults)

	// Overall summary
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("Performance Insights")
	fmt.Println(strings.Repeat("=", 100))

	printInsights(comboResults, cacheResults)

	fmt.Println("\n" + strings.Repeat("=", 100))
}

func printTableHeader() {
	fmt.Printf("%-40s %15s %15s %15s\n", "Scenario", "Time/op", "Allocs/op", "Bytes/op")
	fmt.Println(strings.Repeat("-", 100))
}

func printTableRow(r BenchmarkResult) {
	fmt.Printf("%-40s %15s %15d %15d\n", r.Name, r.Duration.String(), r.Allocs, r.Bytes)
}

func printSummary(section string, results []BenchmarkResult) {
	if len(results) == 0 {
		return
	}

	var totalTime time.Duration

	var totalAllocs, totalBytes int64

	var fastest, slowest BenchmarkResult

	fastest.Duration = time.Hour
	slowest.Duration = 0

	for _, r := range results {
		totalTime += r.Duration
		totalAllocs += r.Allocs
		totalBytes += r.Bytes

		if r.Duration < fastest.Duration {
			fastest = r
		}

		if r.Duration > slowest.Duration {
			slowest = r
		}
	}

	avgTime := totalTime / time.Duration(len(results))

	fmt.Printf("\n%s Summary:\n", section)
	fmt.Printf("  Fastest: %-30s (%s)\n", fastest.Name, fastest.Duration)
	fmt.Printf("  Slowest: %-30s (%s)\n", slowest.Name, slowest.Duration)
	fmt.Printf("  Average: %s\n", avgTime)

	if slowest.Duration > 0 && fastest.Duration > 0 {
		ratio := float64(slowest.Duration) / float64(fastest.Duration)
		fmt.Printf("  Slowdown factor (slowest/fastest): %.2fx\n", ratio)
	}
}

func printInsights(comboResults, cacheResults []BenchmarkResult) {
	// Find defaults_only for baseline
	var baseline time.Duration

	for _, r := range comboResults {
		if r.Name == "defaults_only" {
			baseline = r.Duration

			break
		}
	}

	if baseline > 0 {
		fmt.Println("\n📊 Config Source Impact (vs defaults_only baseline):")

		for _, r := range comboResults {
			if r.Name == "defaults_only" {
				continue
			}

			ratio := float64(r.Duration) / float64(baseline)
			bar := strings.Repeat("█", int(ratio*10))
			fmt.Printf("  %-30s %6.2fx %s\n", r.Name, ratio, bar)
		}
	}

	// Cache insight
	var cacheHit, cacheMiss time.Duration

	for _, r := range cacheResults {
		if strings.Contains(r.Name, "hit") {
			cacheHit = r.Duration
		}

		if strings.Contains(r.Name, "miss") {
			cacheMiss = r.Duration
		}
	}

	if cacheHit > 0 && cacheMiss > 0 {
		speedup := float64(cacheMiss) / float64(cacheHit)
		fmt.Printf("\n🚀 Cache Speedup: %.0fx faster with cache hit\n", speedup)
	}

	// Recommendations
	fmt.Println("\n💡 Recommendations:")
	fmt.Println("  - Use cache hits where possible (avoid Reload() unless config changes)")
	fmt.Println("  - Minimize env vars for fastest startup")
	fmt.Println("  - Project-only config is faster than global+project combined")
	fmt.Println("  - CLI flags have minimal overhead")
}

func benchmarkCombination(t *testing.T, combo ConfigCombination, iterations int) BenchmarkResult {
	t.Helper()

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(filepath.Join(homeDir, ".klaudiush"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(projectDir, ".klaudiush"), 0o755); err != nil {
		t.Fatal(err)
	}

	if combo.HasGlobal && combo.GlobalConfig != nil {
		writeConfigFile(t, filepath.Join(homeDir, ".klaudiush", "config.toml"), combo.GlobalConfig)
	}

	if combo.HasProject && combo.ProjectConfig != nil {
		writeConfigFile(
			t,
			filepath.Join(projectDir, ".klaudiush", "config.toml"),
			combo.ProjectConfig,
		)
	}

	originalEnv := captureEnv()

	clearKlaudiushEnv()

	if combo.HasEnvVars {
		for k, v := range combo.EnvVars {
			os.Setenv(k, v)
		}
	}

	defer restoreEnv(originalEnv)

	var sources []provider.Source
	if combo.HasFlags {
		sources = append(sources, provider.NewFlagSource(combo.Flags))
	}

	sources = append(sources, provider.NewEnvSource())

	loader := config.NewLoaderWithDirs(homeDir, projectDir)
	sources = append(sources, provider.NewProjectFileSource(loader))
	sources = append(sources, provider.NewGlobalFileSource(loader))

	p := provider.NewProvider(sources...)

	return runTimed(combo.Name, iterations, func() {
		p.Reload()
		_, _ = p.Load()
	})
}

func benchmarkMergeOps(iterations int) []BenchmarkResult {
	enabled := true
	disabled := false

	configs := []*pkgconfig.Config{
		config.DefaultConfig(),
		{Global: &pkgconfig.GlobalConfig{UseSDKGit: &enabled}},
		{
			Validators: &pkgconfig.ValidatorsConfig{
				Git: &pkgconfig.GitConfig{
					Commit: &pkgconfig.CommitValidatorConfig{
						ValidatorConfig: pkgconfig.ValidatorConfig{Enabled: &disabled},
					},
				},
			},
		},
		{
			Validators: &pkgconfig.ValidatorsConfig{
				File: &pkgconfig.FileConfig{
					Markdown: &pkgconfig.MarkdownValidatorConfig{
						ValidatorConfig: pkgconfig.ValidatorConfig{Enabled: &enabled},
					},
				},
			},
		},
	}

	merger := config.NewMerger()

	return []BenchmarkResult{
		runTimed(
			"merge_2_configs",
			iterations,
			func() { _ = merger.Merge(configs[0], configs[1]) },
		),
		runTimed(
			"merge_3_configs",
			iterations,
			func() { _ = merger.Merge(configs[0], configs[1], configs[2]) },
		),
		runTimed("merge_4_configs", iterations, func() { _ = merger.Merge(configs...) }),
	}
}

func benchmarkEnvParsing(t *testing.T, iterations int) []BenchmarkResult {
	t.Helper()

	scenarios := []struct {
		name    string
		envVars map[string]string
	}{
		{name: "no_env_vars", envVars: map[string]string{}},
		{name: "single_env_var", envVars: map[string]string{"KLAUDIUSH_USE_SDK_GIT": "true"}},
		{name: "few_env_vars", envVars: map[string]string{
			"KLAUDIUSH_USE_SDK_GIT":                   "true",
			"KLAUDIUSH_DEFAULT_TIMEOUT":               "15s",
			"KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED": "false",
		}},
		{name: "many_env_vars", envVars: map[string]string{
			"KLAUDIUSH_USE_SDK_GIT":                          "true",
			"KLAUDIUSH_DEFAULT_TIMEOUT":                      "15s",
			"KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED":        "false",
			"KLAUDIUSH_VALIDATORS_GIT_COMMIT_SEVERITY":       "warning",
			"KLAUDIUSH_VALIDATORS_GIT_PUSH_ENABLED":          "true",
			"KLAUDIUSH_VALIDATORS_GIT_ADD_ENABLED":           "false",
			"KLAUDIUSH_VALIDATORS_FILE_MARKDOWN_ENABLED":     "true",
			"KLAUDIUSH_VALIDATORS_FILE_SHELLSCRIPT_ENABLED":  "false",
			"KLAUDIUSH_VALIDATORS_NOTIFICATION_BELL_ENABLED": "true",
		}},
	}

	results := make([]BenchmarkResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		originalEnv := captureEnv()

		clearKlaudiushEnv()

		for k, v := range scenario.envVars {
			os.Setenv(k, v)
		}

		source := provider.NewEnvSource()
		result := runTimed(scenario.name, iterations, func() { _, _ = source.Load() })
		results = append(results, result)

		restoreEnv(originalEnv)
	}

	return results
}

func benchmarkFlagParsing(iterations int) []BenchmarkResult {
	scenarios := []struct {
		name  string
		flags map[string]any
	}{
		{name: "no_flags", flags: map[string]any{}},
		{name: "single_flag", flags: map[string]any{"use-sdk-git": true}},
		{
			name:  "few_flags",
			flags: map[string]any{"use-sdk-git": true, "timeout": "15s", "disable": "commit"},
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

	results := make([]BenchmarkResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		source := provider.NewFlagSource(scenario.flags)
		result := runTimed(scenario.name, iterations, func() { _, _ = source.Load() })
		results = append(results, result)
	}

	return results
}

func benchmarkCachePerf(iterations int) []BenchmarkResult {
	p := provider.NewProvider()
	_, _ = p.Load() // Warm up

	hitResult := runTimed("cache_hit", iterations, func() { _, _ = p.Load() })

	missResult := runTimed("cache_miss", iterations, func() {
		p.Reload()
		_, _ = p.Load()
	})

	return []BenchmarkResult{hitResult, missResult}
}

func benchmarkFileLoadingReport(t *testing.T, iterations int) []BenchmarkResult {
	t.Helper()

	scenarios := []struct {
		name   string
		config *pkgconfig.Config
	}{
		{name: "minimal_config", config: &pkgconfig.Config{}},
		{name: "default_config", config: config.DefaultConfig()},
		{name: "full_config", config: createFullConfig()},
	}

	results := make([]BenchmarkResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		tmpDir := t.TempDir()
		homeDir := filepath.Join(tmpDir, "home")

		if err := os.MkdirAll(filepath.Join(homeDir, ".klaudiush"), 0o755); err != nil {
			t.Fatal(err)
		}

		writeConfigFile(t, filepath.Join(homeDir, ".klaudiush", "config.toml"), scenario.config)

		loader := config.NewLoaderWithDirs(homeDir, tmpDir)
		result := runTimed(scenario.name, iterations, func() { _, _ = loader.LoadGlobal() })
		results = append(results, result)
	}

	return results
}

func benchmarkValidatorToggle(iterations int) []BenchmarkResult {
	scenarios := []struct {
		name    string
		disable string
	}{
		{name: "none_disabled", disable: ""},
		{name: "one_disabled", disable: "commit"},
		{name: "few_disabled", disable: "commit,push,markdown"},
		{name: "many_disabled", disable: "commit,push,add,markdown,shellscript,terraform"},
	}

	results := make([]BenchmarkResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		var flags map[string]any
		if scenario.disable != "" {
			flags = map[string]any{"disable": scenario.disable}
		}

		p := provider.NewProvider(provider.NewFlagSource(flags))
		result := runTimed(scenario.name, iterations, func() {
			p.Reload()
			_, _ = p.Load()
		})
		results = append(results, result)
	}

	return results
}

func benchmarkValidationReport(iterations int) []BenchmarkResult {
	scenarios := []struct {
		name   string
		config *pkgconfig.Config
	}{
		{name: "empty_config", config: &pkgconfig.Config{}},
		{name: "default_config", config: config.DefaultConfig()},
		{name: "full_config", config: createFullConfig()},
	}

	validator := config.NewValidator()
	results := make([]BenchmarkResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		result := runTimed(
			scenario.name,
			iterations,
			func() { _ = validator.Validate(scenario.config) },
		)
		results = append(results, result)
	}

	return results
}

func runTimed(name string, iterations int, fn func()) BenchmarkResult {
	// Warm up
	for range 10 {
		fn()
	}

	start := time.Now()

	for range iterations {
		fn()
	}

	elapsed := time.Since(start)

	return BenchmarkResult{
		Name:     name,
		Duration: elapsed / time.Duration(iterations),
		Allocs:   0, // Would need runtime.MemStats for accurate allocs
		Bytes:    0,
	}
}
