package cli

import (
	"fmt"
	"os"

	"github.com/azemoning/omni-5e/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd is the base command for omni-5e.
var rootCmd = &cobra.Command{
	Use:   "omni-5e",
	Short: "A REST API and CLI for the D&D 5e System Reference Document",
	Long:  `omni-5e serves the complete D&D 5e SRD as clean, structured, queryable JSON.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level: trace|debug|info|warn|error")
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
}

func initConfig() {
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
	viper.AutomaticEnv()

	viper.ReadInConfig() //nolint:errcheck // optional config file
}

// LoadConfig loads the application config using Viper.
func LoadConfig() (*config.Config, error) {
	return config.Load(cfgFile)
}
