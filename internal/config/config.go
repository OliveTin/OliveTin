package config

import ()

type Action struct {
	ID          string
	Title       string
	Icon        string
	Shell       string
	CSS         map[string]string `mapstructure:"omitempty"`
	Timeout     int
	Permissions []PermissionsEntry
	Arguments   []ActionArgument
}

type ActionArgument struct {
	Name    string
	Title   string
	Type    string
	Default string
	Choices []ActionArgumentChoice
}

type ActionArgumentChoice struct {
	Value string
	Title string
}

// Entity represents a "thing" that can have multiple actions associated with it.
// for example, a media player with a start and stop action.
type Entity struct {
	Title   string
	Icon    string
	Actions []Action `mapstructure:"actions"`
	CSS     map[string]string
}

type PermissionsEntry struct {
	Usergroup string
	View      bool
	Exec      bool
}

type DefaultPermissions struct {
	View bool
	Exec bool
}

type UserGroup struct {
	Name    string
	Members []string
}

// Config is the global config used through the whole app.
type Config struct {
	UseSingleHTTPFrontend           bool
	ThemeName                       string
	HideNavigation                  bool
	ListenAddressSingleHTTPFrontend string
	ListenAddressWebUI              string
	ListenAddressRestActions        string
	ListenAddressGrpcActions        string
	ExternalRestAddress             string
	LogLevel                        string
	Actions                         []Action `mapstructure:"actions"`
	Entities                        []Entity `mapstructure:"entities"`
	CheckForUpdates                 bool
	ShowNewVersions                 bool
	Usergroups                      []UserGroup
	DefaultPermissions              DefaultPermissions
}

// DefaultConfig gets a new Config structure with sensible default values.
func DefaultConfig() *Config {
	config := Config{}
	config.UseSingleHTTPFrontend = true
	config.HideNavigation = false
	config.ListenAddressSingleHTTPFrontend = "0.0.0.0:1337"
	config.ListenAddressRestActions = "localhost:1338"
	config.ListenAddressGrpcActions = "localhost:1339"
	config.ListenAddressWebUI = "localhost:1340"
	config.LogLevel = "INFO"
	config.CheckForUpdates = true
	config.ShowNewVersions = true
	config.DefaultPermissions.Exec = true
	config.DefaultPermissions.View = true

	return &config
}
