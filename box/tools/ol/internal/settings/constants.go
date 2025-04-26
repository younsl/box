package settings

const (
	// AppName is the application name used for directory structures.
	AppName = "ol"

	// Config Keys
	KeyTimezone      = "timezone"
	KeyDataDirectory = "dataDirectory"

	// Default Values
	DefaultTimezone = "UTC"

	// File/Directory Names
	ConfigFileNameBase = "config"
	ConfigFileType     = "yaml"
	LogFileNameFormat  = "ol-%d.txt"
)
