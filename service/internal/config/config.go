package config

import (
	"fmt"
)

// Action represents the core functionality of OliveTin - commands that show up
// as buttons in the UI.
type Action struct {
	ID                     string
	Title                  string
	Icon                   string
	Shell                  string
	Exec                   []string
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
	Triggers               []string
	MaxConcurrent          int
	MaxRate                []RateSpec
	Arguments              []ActionArgument
	PopupOnStart           string
	SaveLogs               SaveLogsConfig
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
	RejectNull  bool
	Suggestions map[string]string
}

// ActionArgumentChoice represents a predefined choice for an argument.
type ActionArgumentChoice struct {
	Value string
	Title string
}

// RateSpec allows you to set a max frequency for an action.
type RateSpec struct {
	Limit    int
	Duration string
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
	View bool `mapstructure:"view"`
	Exec bool `mapstructure:"exec"`
	Logs bool `mapstructure:"logs"`
	Kill bool `mapstructure:"kill"`
}

// AccessControlList defines what permissions apply to a user or user group.
type AccessControlList struct {
	Name             string
	AddToEveryAction bool
	MatchUsergroups  []string
	MatchUsernames   []string
	Permissions      PermissionsList
	Policy           ConfigurationPolicy
}

// ConfigurationPolicy defines global settings which are overridden with an ACL.
type ConfigurationPolicy struct {
	ShowDiagnostics bool `mapstructure:"showDiagnostics"`
	ShowLogList     bool `mapstructure:"showLogList"`
}

type PrometheusConfig struct {
	Enabled          bool `mapstructure:"enabled"`
	DefaultGoMetrics bool `mapstructure:"defaultGoMetrics"`
}

// Config is the global config used through the whole app.
type Config struct {
	UseSingleHTTPFrontend           bool                       `mapstructure:"useSingleHTTPFrontend"`
	ThemeName                       string                     `mapstructure:"themeName"`
	ThemeCacheDisabled              bool                       `mapstructure:"themeCacheDisabled"`
	ListenAddressSingleHTTPFrontend string                     `mapstructure:"listenAddressSingleHTTPFrontend"`
	ListenAddressWebUI              string                     `mapstructure:"listenAddressWebUI"`
	ListenAddressRestActions        string                     `mapstructure:"listenAddressRestActions"`
	ListenAddressPrometheus         string                     `mapstructure:"listenAddressPrometheus"`
	ExternalRestAddress             string                     `mapstructure:"externalRestAddress"`
	LogLevel                        string                     `mapstructure:"logLevel"`
	LogDebugOptions                 LogDebugOptions            `mapstructure:"logDebugOptions"`
	LogHistoryPageSize              int64                      `mapstructure:"logHistoryPageSize"`
	Actions                         []*Action                  `mapstructure:"actions"`
	Entities                        []*EntityFile              `mapstructure:"entities"`
	Dashboards                      []*DashboardComponent      `mapstructure:"dashboards"`
	CheckForUpdates                 bool                       `mapstructure:"checkForUpdates"`
	PageTitle                       string                     `mapstructure:"pageTitle"`
	ShowFooter                      bool                       `mapstructure:"showFooter"`
	ShowNavigation                  bool                       `mapstructure:"showNavigation"`
	ShowNewVersions                 bool                       `mapstructure:"showNewVersions"`
	EnableCustomJs                  bool                       `mapstructure:"enableCustomJs"`
	AuthJwtCookieName               string                     `mapstructure:"authJwtCookieName"`
	AuthJwtHeader                   string                     `mapstructure:"authJwtHeader"`
	AuthJwtAud                      string                     `mapstructure:"authJwtAud"`
	AuthJwtDomain                   string                     `mapstructure:"authJwtDomain"`
	AuthJwtCertsURL                 string                     `mapstructure:"authJwtCertsUrl"`
	AuthJwtHmacSecret               string                     `mapstructure:"authJwtHmacSecret"` // mutually exclusive with pub key config fields
	AuthJwtClaimUsername            string                     `mapstructure:"authJwtClaimUsername"`
	AuthJwtClaimUserGroup           string                     `mapstructure:"authJwtClaimUserGroup"`
	AuthJwtPubKeyPath               string                     `mapstructure:"authJwtPubKeyPath"` // will read pub key from file on disk
	AuthHttpHeaderUsername          string                     `mapstructure:"authHttpHeaderUsername"`
	AuthHttpHeaderUserGroup         string                     `mapstructure:"authHttpHeaderUserGroup"`
	AuthHttpHeaderUserGroupSep      string                     `mapstructure:"authHttpHeaderUserGroupSep"`
	AuthLocalUsers                  AuthLocalUsersConfig       `mapstructure:"authLocalUsers"`
	AuthLoginUrl                    string                     `mapstructure:"authLoginUrl"`
	AuthRequireGuestsToLogin        bool                       `mapstructure:"authRequireGuestsToLogin"`
	AuthOAuth2RedirectURL           string                     `mapstructure:"authOAuth2RedirectUrl"`
	AuthOAuth2Providers             map[string]*OAuth2Provider `mapstructure:"authOAuth2Providers"`
	DefaultPermissions              PermissionsList            `mapstructure:"defaultPermissions"`
	DefaultPolicy                   ConfigurationPolicy        `mapstructure:"defaultPolicy"`
	AccessControlLists              []*AccessControlList       `mapstructure:"accessControlLists"`
	WebUIDir                        string                     `mapstructure:"webUIDir"`
	CronSupportForSeconds           bool                       `mapstructure:"cronSupportForSeconds"`
	SectionNavigationStyle          string                     `mapstructure:"sectionNavigationStyle"`
	DefaultPopupOnStart             string                     `mapstructure:"defaultPopupOnStart"`
	InsecureAllowDumpOAuth2UserData bool                       `mapstructure:"insecureAllowDumpOAuth2UserData"`
	InsecureAllowDumpVars           bool                       `mapstructure:"insecureAllowDumpVars"`
	InsecureAllowDumpSos            bool                       `mapstructure:"insecureAllowDumpSos"`
	InsecureAllowDumpActionMap      bool                       `mapstructure:"insecureAllowDumpActionMap"`
	InsecureAllowDumpJwtClaims      bool                       `mapstructure:"insecureAllowDumpJwtClaims"`
	Prometheus                      PrometheusConfig           `mapstructure:"prometheus"`
	SaveLogs                        SaveLogsConfig             `mapstructure:"saveLogs"`
	DefaultIconForActions           string                     `mapstructure:"defaultIconForActions"`
	DefaultIconForDirectories       string                     `mapstructure:"defaultIconForDirectories"`
	DefaultIconForBack              string                     `mapstructure:"defaultIconForBack"`
	AdditionalNavigationLinks       []*NavigationLink          `mapstructure:"additionalNavigationLinks"`
	ServiceHostMode                 string                     `mapstructure:"serviceHostMode"`
	StyleMods                       []string                   `mapstructure:"styleMods"`
	BannerMessage                   string                     `mapstructure:"bannerMessage"`
	BannerCSS                       string                     `mapstructure:"bannerCss"`
	Include                         string                     `mapstructure:"include"`

	sourceFiles []string
}

type AuthLocalUsersConfig struct {
	Enabled bool         `mapstructure:"enabled"`
	Users   []*LocalUser `mapstructure:"users"`
}

type LocalUser struct {
	Username  string `mapstructure:"username"`
	Usergroup string `mapstructure:"usergroup"`
	Password  string `mapstructure:"password"`
}

type OAuth2Provider struct {
	Name               string
	Title              string
	ClientID           string
	ClientSecret       string
	Icon               string
	Scopes             []string
	AuthUrl            string
	TokenUrl           string
	WhoamiUrl          string
	UsernameField      string
	UserGroupField     string
	InsecureSkipVerify bool
	CallbackTimeout    int
	CertBundlePath     string
}

type NavigationLink struct {
	Title  string `mapstructure:"title"`
	Url    string `mapstructure:"url"`
	Target string `mapstructure:"target"`
}

type SaveLogsConfig struct {
	ResultsDirectory string `mapstructure:"resultsDirectory"`
	OutputDirectory  string `mapstructure:"outputDirectory"`
}

type LogDebugOptions struct {
	SingleFrontendRequests       bool
	SingleFrontendRequestHeaders bool
	AclCheckStarted              bool
	AclMatched                   bool
	AclNotMatched                bool
	AclNoneMatched               bool
}

type DashboardComponent struct {
	Title    string
	Type     string
	Entity   string
	Icon     string
	CssClass string
	Contents []*DashboardComponent
}

func DefaultConfig() *Config {
	return DefaultConfigWithBasePort(1337)
}

// DefaultConfig gets a new Config structure with sensible default values.
func DefaultConfigWithBasePort(basePort int) *Config {
	config := Config{}
	config.UseSingleHTTPFrontend = true
	config.PageTitle = "OliveTin"
	config.ShowFooter = true
	config.ShowNavigation = true
	config.ShowNewVersions = true
	config.EnableCustomJs = false
	config.ExternalRestAddress = "."
	config.LogLevel = "INFO"
	config.LogHistoryPageSize = 10
	config.CheckForUpdates = false
	config.DefaultPermissions.Exec = true
	config.DefaultPermissions.View = true
	config.DefaultPermissions.Logs = true
	config.DefaultPermissions.Kill = true
	config.AuthJwtClaimUsername = "name"
	config.AuthJwtClaimUserGroup = "group"
	config.AuthRequireGuestsToLogin = false
	config.WebUIDir = "./webui"
	config.CronSupportForSeconds = false
	config.SectionNavigationStyle = "sidebar"
	config.DefaultPopupOnStart = "nothing"
	config.InsecureAllowDumpVars = false
	config.InsecureAllowDumpSos = false
	config.InsecureAllowDumpActionMap = false
	config.InsecureAllowDumpJwtClaims = false
	config.Prometheus.Enabled = false
	config.Prometheus.DefaultGoMetrics = false
	config.DefaultIconForActions = "&#x1F600;"
	config.DefaultIconForDirectories = "&#128193"
	config.DefaultIconForBack = "&laquo;"
	config.ThemeCacheDisabled = false
	config.ServiceHostMode = ""

	config.ListenAddressSingleHTTPFrontend = fmt.Sprintf("0.0.0.0:%d", basePort)
	config.ListenAddressRestActions = fmt.Sprintf("localhost:%d", basePort+1)
	config.ListenAddressWebUI = fmt.Sprintf("localhost:%d", basePort+3)
	config.ListenAddressPrometheus = fmt.Sprintf("localhost:%d", basePort+4)

	config.DefaultPolicy.ShowDiagnostics = true
	config.DefaultPolicy.ShowLogList = true

	return &config
}
