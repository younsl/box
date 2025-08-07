package tui

// AppConfig holds configuration for the TUI application
type AppConfig struct {
	ServerURL   string
	Org         string
	Repo        string
	Environment string
	Token       string
	Timezone    string
	Version     string
}