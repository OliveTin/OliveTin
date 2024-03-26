package config

// Action represents the core functionality of OliveTin - commands that show up
// as buttons in the UI.
type Action struct {
	ID                     string
	Title                  string
	Icon                   string
	Shell                  string
	ShellAfterCompleted    string
	Timeout                int
	Acls                   []string
	Entity                 string
	Hidden                 bool
	ExecOnStartup          bool
	ExecOnCron             []string
	ExecOnFileCreatedInDir []string
	ExecOnFileChangedInDir []string
	ExecOnCalendarFile     string
	Trigger                string
	MaxConcurrent          int
	Arguments              []ActionArgument
	PopupOnStart           string
}

// ActionArgument objects appear on Actions.
type ActionArgument struct {
	Name        string
	Title       string
	Description string
	Type        string
	Default     string
	Choices     []ActionArgumentChoice
	Entity      string
	Suggestions map[string]string
}

// ActionArgumentChoice represents a predefined choice for an argument.
type ActionArgumentChoice struct {
	Value string
	Title string
}

// Entity represents a "thing" that can have multiple actions associated with it.
// for example, a media player with a start and stop action.
type EntityFile struct {
	File string
	Name string
	Icon string
}

// PermissionsList defines what users can do with an action.
type PermissionsList struct {
	View bool
	Exec bool
}

// AccessControlList defines what permissions apply to a user or user group.
type AccessControlList struct {
	Name             string
	AddToEveryAction bool
	MatchUsergroups  []string
	MatchUsernames   []string
	Permissions      PermissionsList
}

type PrometheusConfig struct {
	Enabled          bool
	DefaultGoMetrics bool
}

// Config is the global config used through the whole app.
type Config struct {
	UseSingleHTTPFrontend           bool
	ThemeName                       string
	ListenAddressSingleHTTPFrontend string
	ListenAddressRestActions        string
	ListenAddressGrpcActions        string
	ListenAddressWebUI              string
	ListenAddressPrometheus         string
	ExternalRestAddress             string
	LogLevel                        string
	Actions                         []*Action             `mapstructure:"actions"`
	Entities                        []*EntityFile         `mapstructure:"entities"`
	Dashboards                      []*DashboardComponent `mapstructure:"dashboards"`
	CheckForUpdates                 bool
	PageTitle                       string
	ShowFooter                      bool
	ShowNavigation                  bool
	ShowNewVersions                 bool
	AuthJwtCookieName               string
	AuthJwtSecret                   string // mutually exclusive with pub key config fields
	AuthJwtClaimUsername            string
	AuthJwtClaimUserGroup           string
	AuthJwtPubKeyPath               string // will read pub key from file on disk
	AuthHttpHeaderUsername          string
	AuthHttpHeaderUserGroup         string
	DefaultPermissions              PermissionsList
	AccessControlLists              []*AccessControlList
	WebUIDir                        string
	CronSupportForSeconds           bool
	SectionNavigationStyle          string
	DefaultPopupOnStart             string
	InsecureAllowDumpVars           bool
	InsecureAllowDumpSos            bool
	InsecureAllowDumpActionMap      bool
	Prometheus                      PrometheusConfig
}

type DashboardComponent struct {
	Title    string
	Type     string
	Entity   string
	Contents []DashboardComponent
}

// DefaultConfig gets a new Config structure with sensible default values.
func DefaultConfig() *Config {
	config := Config{}
	config.UseSingleHTTPFrontend = true
	config.PageTitle = "OliveTin"
	config.ShowFooter = true
	config.ShowNavigation = true
	config.ShowNewVersions = true
	config.ListenAddressSingleHTTPFrontend = "0.0.0.0:1337"
	config.ListenAddressRestActions = "localhost:1338"
	config.ListenAddressGrpcActions = "localhost:1339"
	config.ListenAddressWebUI = "localhost:1340"
	config.ListenAddressPrometheus = "localhost:1341"
	config.ExternalRestAddress = "."
	config.LogLevel = "INFO"
	config.CheckForUpdates = true
	config.DefaultPermissions.Exec = true
	config.DefaultPermissions.View = true
	config.AuthJwtClaimUsername = "name"
	config.AuthJwtClaimUserGroup = "group"
	config.WebUIDir = "./webui"
	config.CronSupportForSeconds = false
	config.SectionNavigationStyle = "sidebar"
	config.DefaultPopupOnStart = "nothing"
	config.InsecureAllowDumpVars = false
	config.InsecureAllowDumpSos = false
	config.InsecureAllowDumpActionMap = false
	config.Prometheus.Enabled = false
	config.Prometheus.DefaultGoMetrics = false

	return &config
}
