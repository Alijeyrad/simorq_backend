package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alijeyrad/simorq_backend/pkg/constants"
	"github.com/spf13/viper"
)

var GlobalConf *Config

func ReadConfig(configPath string) (*Config, error) {
	viper.SetConfigName(constants.ConfigName)
	viper.SetConfigType(constants.ConfigFormat)
	viper.AddConfigPath(configPath)

	// Allow env vars to override config values.
	// e.g. SIMORQ_DATABASE_HOST overrides database.host
	viper.SetEnvPrefix("SIMORQ")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read the config file (optional in Docker environments)
	if err := viper.ReadInConfig(); err != nil {
		// If config file not found but we have env vars, continue with defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only fail if it's not a "file not found" error
			if os.Getenv("SIMORQ_DATABASE_HOST") == "" {
				return nil, fmt.Errorf("error reading config file: %v", err)
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %v", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

func MustReadConfig(path string) *Config {
	config, err := ReadConfig(path)
	if err != nil {
		panic(err)
	}

	GlobalConf = config

	return config
}
