package settings

import (
	"github.com/spf13/viper"
)

// ConfigItem holds metadata about a specific configuration setting.
// Note: We use Viper for default handling, so DefaultValue is mostly for reference.
type ConfigItem struct {
	Key         string // Viper key (e.g., KeyTimezone)
	Description string // User-friendly description
	DefaultInfo string // Description of the default behavior or value
	IsPath      bool   // True if the value represents a directory path
}

// AppSettings defines the list of supported application settings.
// Adding a new setting here will automatically include it in `config get` output.
var AppSettings = []ConfigItem{
	{
		Key:         KeyTimezone,
		Description: "Timezone for timestamps",
		DefaultInfo: "Defaults to UTC",
		IsPath:      false,
	},
	{
		Key:         KeyDataDirectory,
		Description: "Directory for data files (logs)",
		DefaultInfo: "Defaults to XDG standard path ($XDG_DATA_HOME or $HOME/.local/share)",
		IsPath:      true,
	},
	// Add future configuration items here
}

// GetValueSource determines if a value comes from the config file or default.
func GetValueSource(key string, configFileUsed string) string {
	if viper.IsSet(key) && configFileUsed != "" {
		return "(from config)"
	}
	return "(default)"
} 