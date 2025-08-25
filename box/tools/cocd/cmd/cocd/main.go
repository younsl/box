package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/younsl/cocd/pkg/config"
	"github.com/younsl/cocd/pkg/github"
	"github.com/younsl/cocd/pkg/monitor"
	"github.com/younsl/cocd/pkg/tui"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "cocd",
	Short: "GitHub Actions Continuous Deployment Monitor",
	Long: `cocd is a TUI application for monitoring GitHub Actions jobs 
that are waiting for approval in production environments.`,
	Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
	RunE:    run,
}

func init() {
	rootCmd.Flags().StringP("config", "c", "", "config file path")
	rootCmd.Flags().StringP("token", "t", "", "GitHub token")
	rootCmd.Flags().StringP("base-url", "u", "", "GitHub base URL (for GitHub Enterprise)")
	rootCmd.Flags().StringP("org", "o", "", "GitHub organization")
	rootCmd.Flags().StringP("repo", "r", "", "GitHub repository (optional, if not specified monitors all repos in org)")
	rootCmd.Flags().IntP("interval", "i", 5, "Refresh interval in seconds")
}

func run(cmd *cobra.Command, args []string) error {
	// Check if we're running in a terminal, but allow override
	if !isTerminal() && os.Getenv("FORCE_TTY") != "1" {
		// Store warning for TUI instead of printing to stdout
		fmt.Fprintf(os.Stderr, "Warning: Not running in a TTY. Key input may not work properly.\n")
		fmt.Fprintf(os.Stderr, "Try running in a proper terminal, or set FORCE_TTY=1 to override.\n")
		fmt.Fprintf(os.Stderr, "Terminal info: stdin=%t, stdout=%t, stderr=%t\n", 
			term.IsTerminal(int(os.Stdin.Fd())), 
			term.IsTerminal(int(os.Stdout.Fd())), 
			term.IsTerminal(int(os.Stderr.Fd())))
		fmt.Fprintf(os.Stderr, "Environment: TERM=%s, FORCE_TTY=%s\n", 
			os.Getenv("TERM"), os.Getenv("FORCE_TTY"))
		fmt.Fprintf(os.Stderr, "Continuing anyway...\n")
	}
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if token, _ := cmd.Flags().GetString("token"); token != "" {
		cfg.GitHub.Token = token
	}
	if baseURL, _ := cmd.Flags().GetString("base-url"); baseURL != "" {
		cfg.GitHub.BaseURL = baseURL
	}
	if org, _ := cmd.Flags().GetString("org"); org != "" {
		cfg.GitHub.Org = org
	}
	if repo, _ := cmd.Flags().GetString("repo"); repo != "" {
		cfg.GitHub.Repo = repo
	}
	if interval, _ := cmd.Flags().GetInt("interval"); interval != 0 {
		cfg.Monitor.Interval = interval
	}

	if cfg.GitHub.Org == "" {
		return fmt.Errorf("GitHub organization is required")
	}

	var client *github.Client
	if cfg.GitHub.Repo != "" {
		client, err = github.NewClient(
			cfg.GitHub.Token,
			cfg.GitHub.BaseURL,
			cfg.GitHub.Org,
			cfg.GitHub.Repo,
		)
	} else {
		client, err = github.NewClient(
			cfg.GitHub.Token,
			cfg.GitHub.BaseURL,
			cfg.GitHub.Org,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	mon := monitor.NewMonitor(client, cfg.Monitor.Interval)
	
	tuiConfig := &tui.AppConfig{
		ServerURL:   cfg.GitHub.BaseURL,
		Org:         cfg.GitHub.Org,
		Repo:        cfg.GitHub.Repo,
		Timezone:    cfg.Monitor.Timezone,
		Version:     version,
	}
	
	// Use Bubble Tea instead of tview for better key handling
	monitorAdapter := tui.NewMonitorAdapter(mon)
	if err := tui.RunBubbleApp(monitorAdapter, tuiConfig); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		return fmt.Errorf("failed to run application: %w", err)
	}

	return nil
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}