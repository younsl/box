package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/younsl/kk/internal/checker"
	"github.com/younsl/kk/internal/config"
	"github.com/younsl/kk/internal/logger"
)

var (
	version    = "0.0.1"
	configPath string
)

var rootCmd = &cobra.Command{
	Use:     "kk",
	Short:   "kk checks domain configurations",
	Long:    `kk validates domain configurations based on a provided YAML file. It checks various aspects of the domain setup to ensure correctness and adherence to standards.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		defer logger.Sync()

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			logger.Log.Errorf("Failed to load config '%s': %v", configPath, err)
			return err
		}

		logger.Log.Infof("Loaded domain list from '%s'", configPath)
		checker.RunChecks(cfg.Domains)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the YAML configuration file (e.g., configs/domains.yaml)")
	rootCmd.MarkPersistentFlagRequired("config")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
