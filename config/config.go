// Package config loads goproject configuration from defaults, config files,
// and environment variables.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// EnvPrefix is the prefix for environment variable overrides. A key such as
// log.level is read from GOPROJECT_LOG_LEVEL.
const EnvPrefix = "GOPROJECT"

// Config represents the goproject configuration.
// Its zero value is not useful; use [Defaults] or [Load] to initialize it.
type Config struct {
	Log LogConfig `mapstructure:"log" yaml:"log"`
	// Quiet suppresses all output except errors.
	Quiet bool `mapstructure:"quiet" yaml:"quiet"`
	// Verbose enables debug output.
	Verbose bool `mapstructure:"verbose" yaml:"verbose"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	// Level is the minimum log level: debug, info, warn, or error.
	Level string `mapstructure:"level" yaml:"level"`
	// Format is the log output format: text, json, or color.
	Format string `mapstructure:"format" yaml:"format"`
}

// Defaults returns a config populated with the values from [SetDefaults].
func Defaults() *Config {
	v := viper.New()
	SetDefaults(v)
	cfg := &Config{}
	_ = v.Unmarshal(cfg)
	return cfg
}

// SetDefaults registers every default value on a Viper instance. It is the
// single source of truth for defaults.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("quiet", false)
	v.SetDefault("verbose", false)
}

// Load searches the standard config directories (see [GetConfigDirs]) and the
// current directory for config.yaml. A missing file is not an error.
// It returns the resolved Config and the Viper instance that produced it,
// so callers can bind flags without recreating the precedence chain.
func Load() (*Config, *viper.Viper, error) {
	return load(func(v *viper.Viper) error {
		v.SetConfigName("config")
		for _, dir := range GetConfigDirs() {
			v.AddConfigPath(dir)
		}
		v.AddConfigPath(".")

		var notFound viper.ConfigFileNotFoundError
		if err := v.ReadInConfig(); err != nil && !errors.As(err, &notFound) {
			return err
		}
		return nil
	})
}

// LoadFromPath loads config from an explicit file path. Unlike [Load], a
// missing file is an error.
func LoadFromPath(path string) (*Config, *viper.Viper, error) {
	return load(func(v *viper.Viper) error {
		v.SetConfigFile(path)
		return v.ReadInConfig()
	})
}

func load(read func(*viper.Viper) error) (*Config, *viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	SetDefaults(v)

	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := read(v); err != nil {
		return nil, nil, fmt.Errorf("config load failed: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, nil, fmt.Errorf("config unmarshal failed: %w", err)
	}

	return cfg, v, nil
}
