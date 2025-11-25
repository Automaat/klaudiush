package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

const shortCommitLength = 12

// Build information set by ldflags at build time.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version and build information for klaudiush.",
	Run:   runVersion,
}

// versionRequested is set by the --version/-v flag.
var versionRequested bool

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolVarP(
		&versionRequested,
		"version",
		"v",
		false,
		"Print version information",
	)
}

func checkVersionFlag() {
	if versionRequested {
		fmt.Print(versionString())
		os.Exit(0)
	}
}

func runVersion(_ *cobra.Command, _ []string) {
	fmt.Print(versionString())
}

func versionString() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("klaudiush %s\n", version))
	b.WriteString(fmt.Sprintf("  commit:    %s\n", commit))
	b.WriteString(fmt.Sprintf("  built:     %s\n", date))
	b.WriteString(fmt.Sprintf("  go:        %s\n", runtime.Version()))
	b.WriteString(fmt.Sprintf("  os/arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH))

	if info, ok := debug.ReadBuildInfo(); ok {
		b.WriteString(fmt.Sprintf("  module:    %s\n", info.Main.Path))

		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				if commit == "unknown" {
					b.WriteString(fmt.Sprintf(
						"  vcs.rev:   %s\n",
						setting.Value[:min(shortCommitLength, len(setting.Value))],
					))
				}
			}

			if setting.Key == "vcs.modified" && setting.Value == "true" {
				b.WriteString("  modified:  true\n")
			}
		}
	}

	return b.String()
}
