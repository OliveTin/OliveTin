package config

import (
	"fmt"
)

// Action represents the core functionality of OliveTin - commands that show up
// as buttons in the UI.
type Action struct {
	ID                     string           `koanf:"id"`
	Title                  string           `koanf:"title"`
	Icon                   string           `koanf:"icon"`
	Shell                  string           `koanf:"shell"`
	Exec                   []string         `koanf:"exec"`
	ShellAfterCompleted    string           `koanf:"shellAfterCompleted"`
	Timeout                int              `koanf:"timeout"`
	Acls                   []string         `koanf:"acls"`
	Entity                 string           `koanf:"entity"`
	Hidden                 bool             `koanf:"hidden"`
	ExecOnStartup          bool             `koanf:"execOnStartup"`
	ExecOnCron             []string         `koanf:"execOnCron"`
	ExecOnFileCreatedInDir []string         `koanf:"execOnFileCreatedInDir"`
	ExecOnFileChangedInDir []string         `koanf:"execOnFileChangedInDir"`
	ExecOnCalendarFile     string           `koanf:"execOnCalendarFile"`
	ExecOnWebhook          []WebhookConfig  `koanf:"execOnWebhook"`
	Triggers               []string         `koanf:"triggers"`
	MaxConcurrent          int              `koanf:"maxConcurrent"`
	MaxRate                []RateSpec       `koanf:"maxRate"`
	Arguments              []ActionArgument `koanf:"arguments"`
	PopupOnStart           string           `koanf:"popupOnStart"`
	SaveLogs               SaveLogsConfig   `koanf:"saveLogs"`
	EnabledExpression      string           `koanf:"enabledExpression"`
}

// ActionArgument objects appear on Actions.
type ActionArgument struct {
	Name                  string                 `koanf:"name"`
	Title                 string                 `koanf:"title"`
	Description           string                 `koanf:"description"`
	Type                  string                 `koanf:"type"`
	Default               string                 `koanf:"default"`
	Choices               []ActionArgumentChoice `koanf:"choices"`
	Entity                string                 `koanf:"entity"`
	RejectNull            bool                   `koanf:"rejectNull"`
	Suggestions           map[string]string      `koanf:"suggestions"`
	SuggestionsBrowserKey string                 `koanf:"suggestionsBrowserKey"`
}

// ActionArgumentChoice represents a predefined choice for an argument.
type ActionArgumentChoice struct {
	Value string `koanf:"value"`
	Title string `koanf:"title"`
}

// RateSpec allows you to set a max frequency for an action.
type RateSpec struct {
	Limit    int    `koanf:"limit"`
	Duration string `koanf:"duration"`
}

// WebhookConfig defines configuration for generic webhook triggers.
type WebhookConfig struct {
	Secret       string            `koanf:"secret"`       // Optional: secret for signature verification
	AuthType     string            `koanf:"authType"`     // Optional: "hmac-sha256", "hmac-sha1", "bearer", "basic", "none"
	AuthHeader   string            `koanf:"authHeader"`   // Optional: custom header name for auth (default: "X-Webhook-Signature")
	MatchHeaders map[string]string `koanf:"matchHeaders"` // Match HTTP headers
	MatchPath    string            `koanf:"matchPath"`    // JSONPath expression to match in request body (format: "jsonpath=value" or just "jsonpath")
	MatchQuery   map[string]string `koanf:"matchQuery"`   // Match URL query parameters
	Extract      map[string]string `koanf:"extract"`      // Map action argument names to JSONPath expressions
	Template     string            `koanf:"template"`     // Optional: template name (e.g., "github-push", "github-pr")
}

// Entity represents a "thing" that can have multiple actions associated with it.
// for example, a media player with a start and stop action.
type EntityFile struct {
	File string `koanf:"file"`
	Name string `koanf:"name"`
	Icon string `koanf:"icon"`
}

// PermissionsList defines what users can do with an action.
type PermissionsList struct {
	View bool `koanf:"view"`
	Exec bool `koanf:"exec"`
	Logs bool `koanf:"logs"`
	Kill bool `koanf:"kill"`
}

// AccessControlList defines what permissions apply to a user or user group.
type AccessControlList struct {
	Name             string              `koanf:"name"`
	AddToEveryAction bool                `koanf:"addToEveryAction"`
	MatchUsergroups  []string            `koanf:"matchUsergroups"`
	MatchUsernames   []string            `koanf:"matchUsernames"`
	Permissions      PermissionsList     `koanf:"permissions"`
	Policy           ConfigurationPolicy `koanf:"policy"`
}

// ConfigurationPolicy defines global settings which are overridden with an ACL.
type ConfigurationPolicy struct {
	ShowDiagnostics bool `koanf:"showDiagnostics"`
	ShowLogList     bool `koanf:"showLogList"`
}

type PrometheusConfig struct {
	Enabled          bool `koanf:"enabled"`
	DefaultGoMetrics bool `koanf:"defaultGoMetrics"`
}

// Config is the global config used through the whole app.
type Config struct {
	UseSingleHTTPFrontend           bool                       `koanf:"useSingleHTTPFrontend"`
	ThemeName                       string                     `koanf:"themeName"`
	ThemeCacheDisabled              bool                       `koanf:"themeCacheDisabled"`
	ListenAddressSingleHTTPFrontend string                     `koanf:"listenAddressSingleHTTPFrontend"`
	ListenAddressWebUI              string                     `koanf:"listenAddressWebUI"`
	ListenAddressRestActions        string                     `koanf:"listenAddressRestActions"`
	ListenAddressPrometheus         string                     `koanf:"listenAddressPrometheus"`
	ExternalRestAddress             string                     `koanf:"externalRestAddress"`
	LogLevel                        string                     `koanf:"logLevel"`
	LogDebugOptions                 LogDebugOptions            `koanf:"logDebugOptions"`
	LogHistoryPageSize              int64                      `koanf:"logHistoryPageSize"`
	Actions                         []*Action                  `koanf:"actions"`
	Entities                        []*EntityFile              `koanf:"entities"`
	Dashboards                      []*DashboardComponent      `koanf:"dashboards"`
	CheckForUpdates                 bool                       `koanf:"checkForUpdates"`
	PageTitle                       string                     `koanf:"pageTitle"`
	ShowFooter                      bool                       `koanf:"showFooter"`
	ShowNavigation                  bool                       `koanf:"showNavigation"`
	ShowNewVersions                 bool                       `koanf:"showNewVersions"`
	ShowNavigateOnStartIcons        bool                       `koanf:"showNavigateOnStartIcons"`
	EnableCustomJs                  bool                       `koanf:"enableCustomJs"`
	AuthJwtCookieName               string                     `koanf:"authJwtCookieName"`
	AuthJwtHeader                   string                     `koanf:"authJwtHeader"`
	AuthJwtAud                      string                     `koanf:"authJwtAud"`
	AuthJwtDomain                   string                     `koanf:"authJwtDomain"`
	AuthJwtCertsURL                 string                     `koanf:"authJwtCertsUrl"`
	AuthJwtHmacSecret               string                     `koanf:"authJwtHmacSecret"` // mutually exclusive with pub key config fields
	AuthJwtClaimUsername            string                     `koanf:"authJwtClaimUsername"`
	AuthJwtClaimUserGroup           string                     `koanf:"authJwtClaimUserGroup"`
	AuthJwtPubKeyPath               string                     `koanf:"authJwtPubKeyPath"` // will read pub key from file on disk
	AuthHttpHeaderUsername          string                     `koanf:"authHttpHeaderUsername"`
	AuthHttpHeaderUserGroup         string                     `koanf:"authHttpHeaderUserGroup"`
	AuthHttpHeaderUserGroupSep      string                     `koanf:"authHttpHeaderUserGroupSep"`
	AuthLocalUsers                  AuthLocalUsersConfig       `koanf:"authLocalUsers"`
	AuthLoginUrl                    string                     `koanf:"authLoginUrl"`
	AuthRequireGuestsToLogin        bool                       `koanf:"authRequireGuestsToLogin"`
	AuthOAuth2RedirectURL           string                     `koanf:"authOAuth2RedirectUrl"`
	AuthOAuth2Providers             map[string]*OAuth2Provider `koanf:"authOAuth2Providers"`
	DefaultPermissions              PermissionsList            `koanf:"defaultPermissions"`
	DefaultPolicy                   ConfigurationPolicy        `koanf:"defaultPolicy"`
	AccessControlLists              []*AccessControlList       `koanf:"accessControlLists"`
	WebUIDir                        string                     `koanf:"webUIDir"`
	CronSupportForSeconds           bool                       `koanf:"cronSupportForSeconds"`
	SectionNavigationStyle          string                     `koanf:"sectionNavigationStyle"`
	DefaultPopupOnStart             string                     `koanf:"defaultPopupOnStart"`
	InsecureAllowDumpOAuth2UserData bool                       `koanf:"insecureAllowDumpOAuth2UserData"`
	InsecureAllowDumpVars           bool                       `koanf:"insecureAllowDumpVars"`
	InsecureAllowDumpSos            bool                       `koanf:"insecureAllowDumpSos"`
	InsecureAllowDumpActionMap      bool                       `koanf:"insecureAllowDumpActionMap"`
	InsecureAllowDumpJwtClaims      bool                       `koanf:"insecureAllowDumpJwtClaims"`
	Prometheus                      PrometheusConfig           `koanf:"prometheus"`
	SaveLogs                        SaveLogsConfig             `koanf:"saveLogs"`
	DefaultIconForActions           string                     `koanf:"defaultIconForActions"`
	DefaultIconForDirectories       string                     `koanf:"defaultIconForDirectories"`
	DefaultIconForBack              string                     `koanf:"defaultIconForBack"`
	AdditionalNavigationLinks       []*NavigationLink          `koanf:"additionalNavigationLinks"`
	ServiceHostMode                 string                     `koanf:"serviceHostMode"`
	StyleMods                       []string                   `koanf:"styleMods"`
	BannerMessage                   string                     `koanf:"bannerMessage"`
	BannerCSS                       string                     `koanf:"bannerCss"`
	Include                         string                     `koanf:"include"`

	sourceFiles []string
}

type AuthLocalUsersConfig struct {
	Enabled bool         `koanf:"enabled"`
	Users   []*LocalUser `koanf:"users"`
}

type LocalUser struct {
	Username  string `koanf:"username"`
	Usergroup string `koanf:"usergroup"`
	Password  string `koanf:"password"`
}

type OAuth2Provider struct {
	Name               string   `koanf:"name"`
	Title              string   `koanf:"title"`
	ClientID           string   `koanf:"clientId"`
	ClientSecret       string   `koanf:"clientSecret"`
	Icon               string   `koanf:"icon"`
	Scopes             []string `koanf:"scopes"`
	AuthUrl            string   `koanf:"authUrl"`
	TokenUrl           string   `koanf:"tokenUrl"`
	WhoamiUrl          string   `koanf:"whoamiUrl"`
	UsernameField      string   `koanf:"usernameField"`
	UserGroupField     string   `koanf:"userGroupField"`
	InsecureSkipVerify bool     `koanf:"insecureSkipVerify"`
	CallbackTimeout    int      `koanf:"callbackTimeout"`
	CertBundlePath     string   `koanf:"certBundlePath"`
	AddToUsergroup     string   `koanf:"addToUsergroup"`
}

type NavigationLink struct {
	Title  string `koanf:"title"`
	Url    string `koanf:"url"`
	Target string `koanf:"target"`
}

type SaveLogsConfig struct {
	ResultsDirectory string `koanf:"resultsDirectory"`
	OutputDirectory  string `koanf:"outputDirectory"`
}

type LogDebugOptions struct {
	SingleFrontendRequests       bool `koanf:"singleFrontendRequests"`
	SingleFrontendRequestHeaders bool `koanf:"singleFrontendRequestHeaders"`
	AclCheckStarted              bool `koanf:"aclCheckStarted"`
	AclMatched                   bool `koanf:"aclMatched"`
	AclNotMatched                bool `koanf:"aclNotMatched"`
	AclNoneMatched               bool `koanf:"aclNoneMatched"`
}

type DashboardComponent struct {
	Title        string                `koanf:"title"`
	Type         string                `koanf:"type"`
	Entity       string                `koanf:"entity"`
	Icon         string                `koanf:"icon"`
	CssClass     string                `koanf:"cssClass"`
	InlineAction *Action               `koanf:"inlineAction"`
	Contents     []*DashboardComponent `koanf:"contents"`
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
	config.ShowNavigateOnStartIcons = true
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
