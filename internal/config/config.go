package config

import ()

// ActionButton represents a button that is shown in the webui.
type ActionButton struct {
	ID      string
	Title   string
	Icon    string
	Shell   string
	CSS     map[string]string `mapstructure:"omitempty"`
	Timeout int
}

// Entity represents a "thing" that can have multiple actions associated with it.
// for example, a media player with a start and stop action.
type Entity struct {
	Title         string
	Icon          string
	ActionButtons []ActionButton `mapstructure:"actions"`
	CSS           map[string]string
}

// Config is the global config used through the whole app.
type Config struct {
	UseSingleHTTPFrontend           bool
	ThemeName						string
	ListenAddressSingleHTTPFrontend string
	ListenAddressWebUI              string
	ListenAddressRestActions        string
	ListenAddressGrpcActions        string
	ExternalRestAddress             string
	LogLevel                        string
	ActionButtons                   []ActionButton `mapstructure:"actions"`
	Entities                        []Entity       `mapstructure:"omitempty"`
	CheckForUpdates                 bool
}

// DefaultConfig gets a new Config structure with sensible default values.
func DefaultConfig() *Config {
	config := Config{}
	config.UseSingleHTTPFrontend = true
	config.ListenAddressSingleHTTPFrontend = "0.0.0.0:1337"
	config.ListenAddressRestActions = "localhost:1338"
	config.ListenAddressGrpcActions = "localhost:1339"
	config.ListenAddressWebUI = "localhost:1340"
	config.LogLevel = "INFO"
	config.CheckForUpdates = true

	return &config
}
