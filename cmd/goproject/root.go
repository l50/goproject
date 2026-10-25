package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/l50/goproject/config"
	"github.com/l50/goproject/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configKeyType struct{}

var configKey = configKeyType{}

// NewRootCmd constructs the root cobra command for the goproject CLI.
// It wires subcommands, persistent flags, and completion handlers.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "goproject",
		Short: "A starting point for new Go CLIs",
		Long: `goproject is a template for Go command line tools.

Configuration precedence (highest to lowest):
  1. CLI flags (--log-level, --log-format, ...)
  2. Environment variables (GOPROJECT_LOG_LEVEL, GOPROJECT_LOG_FORMAT, ...)
  3. Configuration file ($XDG_CONFIG_HOME/goproject/config.yaml)
  4. Built-in defaults`,
		Version:           version,
		Args:              cobra.NoArgs,
		SilenceUsage:      true,
		PersistentPreRunE: initConfig,
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is $HOME/.config/goproject/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "", "Log format (text, json, color)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode - only show errors")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose mode - show debug output")

	_ = rootCmd.RegisterFlagCompletionFunc("log-level", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = rootCmd.RegisterFlagCompletionFunc("log-format", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"text", "json", "color"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newCompletionCmd())

	return rootCmd
}

// Execute runs the root command.
// It installs signal handling so the CLI exits cleanly on interruption.
func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return NewRootCmd().ExecuteContext(ctx)
}

func initConfig(cmd *cobra.Command, _ []string) error {
	// Read the flag from the command rather than a package-level variable so
	// each NewRootCmd starts from clean state.
	configPath, err := cmd.Root().PersistentFlags().GetString("config")
	if err != nil {
		return fmt.Errorf("failed to read config flag: %w", err)
	}

	var (
		cfg *config.Config
		v   *viper.Viper
	)
	if configPath != "" {
		cfg, v, err = config.LoadFromPath(configPath)
	} else {
		cfg, v, err = config.Load()
	}
	if err != nil {
		// Load only fails when a config file exists but cannot be read or
		// parsed. Surface that rather than silently using defaults.
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := bindPersistentFlags(cmd, v); err != nil {
		return err
	}

	// Re-unmarshal so flag overrides reach the config struct.
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to apply config overrides: %w", err)
	}

	if cfg.Quiet && cfg.Verbose {
		return errors.New("--quiet and --verbose cannot be used together")
	}

	logger := logging.New(cfg.Log.Level, cfg.Log.Format, cfg.Quiet, cfg.Verbose)
	logging.SetDefault(logger)

	ctx := context.WithValue(cmd.Context(), configKey, cfg)
	cmd.SetContext(logging.WithLogger(ctx, logger))

	return nil
}

// bindPersistentFlags binds root persistent flags to their Viper keys.
func bindPersistentFlags(cmd *cobra.Command, v *viper.Viper) error {
	bindings := []struct {
		key  string
		flag string
	}{
		{"log.level", "log-level"},
		{"log.format", "log-format"},
		{"quiet", "quiet"},
		{"verbose", "verbose"},
	}
	for _, b := range bindings {
		if err := v.BindPFlag(b.key, cmd.Root().PersistentFlags().Lookup(b.flag)); err != nil {
			return fmt.Errorf("failed to bind %s flag: %w", b.flag, err)
		}
	}
	return nil
}

// configFromContext returns the config initConfig stored on the command
// context, or nil when it has not run.
func configFromContext(cmd *cobra.Command) *config.Config {
	if cfg, ok := cmd.Context().Value(configKey).(*config.Config); ok {
		return cfg
	}
	return nil
}
