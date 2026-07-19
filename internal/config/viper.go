package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration from env vars, config file, and flags.
func Load(cfgFile string) (*Config, error) {
	cfg := DefaultConfig()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.omni-5e")
		viper.AddConfigPath("/etc/omni-5e")
	}

	viper.SetEnvPrefix("OMNI5E")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Server defaults
	viper.SetDefault("server.host", cfg.Server.Host)
	viper.SetDefault("server.port", cfg.Server.Port)
	viper.SetDefault("server.read_timeout", cfg.Server.ReadTimeout)
	viper.SetDefault("server.write_timeout", cfg.Server.WriteTimeout)
	viper.SetDefault("server.idle_timeout", cfg.Server.IdleTimeout)

	// Database defaults
	viper.SetDefault("database.host", cfg.Database.Host)
	viper.SetDefault("database.port", cfg.Database.Port)
	viper.SetDefault("database.user", cfg.Database.User)
	viper.SetDefault("database.password", cfg.Database.Password)
	viper.SetDefault("database.name", cfg.Database.Name)
	viper.SetDefault("database.sslmode", cfg.Database.SSLMode)
	viper.SetDefault("database.max_conns", cfg.Database.MaxConns)

	// Log defaults
	viper.SetDefault("log.level", cfg.Log.Level)
	viper.SetDefault("log.format", cfg.Log.Format)

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return cfg, nil
}
