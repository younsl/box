package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/younsl/kk/internal/checker"
	"github.com/younsl/kk/internal/config"
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
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error loading config: %v\n", err)
			return fmt.Errorf("failed to load config '%s': %w", configPath, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Loaded domain list from '%s'.\n", configPath)
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
