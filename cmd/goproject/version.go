package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Overridden at release time via -ldflags (see .goreleaser.yaml).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	if commit == "none" || date == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if commit == "none" && len(setting.Value) > 7 {
						commit = setting.Value[:7]
					}
				case "vcs.time":
					if date == "unknown" {
						date = setting.Value
					}
				}
			}
		}
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		Long:    "Display the version, commit hash, and build date of goproject.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(),
				"goproject version %s\n  commit: %s\n  built:  %s\n",
				version, commit, date)
			return err
		},
	}
}
