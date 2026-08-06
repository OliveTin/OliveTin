package config

import (
	"fmt"
)

// ReservedArgumentNamePrefix is reserved for OliveTin-injected system arguments.
const ReservedArgumentNamePrefix = "ot_"

// JustificationRequiredNoTemplate requires a manual justification with no prefilled template.
const JustificationRequiredNoTemplate = " "

// Action represents the core functionality of OliveTin - commands that show up
// as buttons in the UI.
type Action struct {
	SaveLogs               SaveLogsConfig   `koanf:"saveLogs"`
	Shell                  string           `koanf:"shell"`
	Icon                   string           `koanf:"icon"`
	ExecOnCalendarFile     string           `koanf:"execOnCalendarFile"`
	SourceFile             string           `koanf:"-"`
	ShellAfterCompleted    string           `koanf:"shellAfterCompleted"`
	Justification          string           `koanf:"justification"`
	EnabledExpression      string           `koanf:"enabledExpression"`
	Entity                 string           `koanf:"entity"`
	Title                  string           `koanf:"title"`
	PopupOnStart           string           `koanf:"popupOnStart"`
	OnClick                string           `koanf:"onclick"`
	ID                     string           `koanf:"id"`
	MaxRate                []RateSpec       `koanf:"maxRate"`
	Acls                   []string         `koanf:"acls"`
	ExecOnWebhook          []WebhookConfig  `koanf:"execOnWebhook"`
	Triggers               []string         `koanf:"triggers"`
	Exec                   []string         `koanf:"exec"`
	ExecOnFileCreatedInDir []string         `koanf:"execOnFileCreatedInDir"`
	Arguments              []ActionArgument `koanf:"arguments"`
	ExecOnCron             []string         `koanf:"execOnCron"`
	Groups                 []string         `koanf:"groups"`
	ExecOnFileChangedInDir []string         `koanf:"execOnFileChangedInDir"`
	Timeout                int              `koanf:"timeout"`
	MaxConcurrent          int              `koanf:"maxConcurrent"`
	Hidden                 bool             `koanf:"hidden"`
	ExecOnStartup          bool             `koanf:"execOnStartup"`
}

func (action *Action) RequiresJustification() bool {
	return action != nil && action.Justification != ""
}

func (action *Action) JustificationTemplateText() string {
	if !action.RequiresJustification() {
		return ""
	}

	if action.Justification == JustificationRequiredNoTemplate {
		return ""
	}

	return action.Justification
}

// ActionGroup defines shared limits and metadata for a set of actions.
type ActionGroup struct {
	Icon          string `koanf:"icon"`
	MaxConcurrent int    `koanf:"maxConcurrent"`
	QueueSize     int    `koanf:"queueSize"`
}

// ActionArgument objects appear on Actions.
type ActionArgument struct {
	Suggestions           map[string]string      `koanf:"suggestions"`
	Name                  string                 `koanf:"name"`
	Title                 string                 `koanf:"title"`
	Description           string                 `koanf:"description"`
	Type                  string                 `koanf:"type"`
	Default               string                 `koanf:"default"`
	Entity                string                 `koanf:"entity"`
	SuggestionsBrowserKey string                 `koanf:"suggestionsBrowserKey"`
	Choices               []ActionArgumentChoice `koanf:"choices"`
	RejectNull            bool                   `koanf:"rejectNull"`
}

// ActionArgumentChoice represents a predefined choice for an argument.
type ActionArgumentChoice struct {
	Value string `koanf:"value"`
	Title string `koanf:"title"`
}

// RateSpec allows you to set a max frequency for an action.
type RateSpec struct {
	Duration string `koanf:"duration"`
	Limit    int    `koanf:"limit"`
}

// WebhookConfig defines configuration for generic webhook triggers.
type WebhookConfig struct {
	Secret        string            `koanf:"secret"`        // Optional: secret for signature verification
	AuthType      string            `koanf:"authType"`      // Optional: "hmac-sha256", "hmac-sha1", "bearer", "basic", "none"
	AuthHeader    string            `koanf:"authHeader"`    // Optional: custom header name for auth (default: "X-Webhook-Signature")
	MatchHeaders  map[string]string `koanf:"matchHeaders"`  // Match HTTP headers
	MatchPath     string            `koanf:"matchPath"`     // JSONPath expression to match in request body (format: "jsonpath=value" or just "jsonpath")
	MatchQuery    map[string]string `koanf:"matchQuery"`    // Match URL query parameters
	Extract       map[string]string `koanf:"extract"`       // Map action argument names to JSONPath expressions
	Template      string            `koanf:"template"`      // Optional: template name (e.g., "github-push", "github-pr")
	Justification string            `koanf:"justification"` // Optional JSONPath to extract justification from webhook body
}

// Entity represents a "thing" that can have multiple actions associated with it.
// for example, a media player with a start and stop action.
type EntityFile struct {
	File       string           `koanf:"file"`
	Name       string           `koanf:"name"`
	Icon       string           `koanf:"icon"`
	SourceFile string           `koanf:"-"`
	Properties []EntityProperty `koanf:"properties"`
	Acls       []string         `koanf:"acls"`
}

// EntityProperty defines a column shown when listing entity instances in the UI.
type EntityProperty struct {
	Name  string `koanf:"name"`
	Title string `koanf:"title"`
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
	MatchUsergroups  []string            `koanf:"matchUsergroups"`
	MatchUsernames   []string            `koanf:"matchUsernames"`
	Permissions      PermissionsList     `koanf:"permissions"`
	Policy           ConfigurationPolicy `koanf:"policy"`
	AddToEveryAction bool                `koanf:"addToEveryAction"`
}

// ConfigurationPolicy defines global settings which are overridden with an ACL.
type ConfigurationPolicy struct {
	ShowDiagnostics   bool `koanf:"showDiagnostics"`
	ShowLogList       bool `koanf:"showLogList"`
	ShowVersionNumber bool `koanf:"showVersionNumber"`
}

// FeaturesConfig holds global opt-in feature flags. New flags default to false.
type FeaturesConfig struct {
	HeaderSearch bool `koanf:"headerSearch"`
}

type PrometheusConfig struct {
	Enabled          bool `koanf:"enabled"`
	DefaultGoMetrics bool `koanf:"defaultGoMetrics"`
}

// SecurityConfig allows users to fine tune the security related HTTP headers and cookie options.
type SecurityConfig struct {
	ContentSecurityPolicy       string `koanf:"contentSecurityPolicy"`
	XFrameOptions               string `koanf:"xFrameOptions"`
	HeaderContentSecurityPolicy bool   `koanf:"headerContentSecurityPolicy"`
	HeaderXContentTypeOptions   bool   `koanf:"headerXContentTypeOptions"`
	HeaderXFrameOptions         bool   `koanf:"headerXFrameOptions"`
	ForceSecureCookies          bool   `koanf:"forceSecureCookies"`
}

// Config is the global config used through the whole app.
type Config struct {
	ActionGroups                       map[string]*ActionGroup    `koanf:"actionGroups"`
	AuthOAuth2Providers                map[string]*OAuth2Provider `koanf:"authOAuth2Providers"`
	SaveLogs                           SaveLogsConfig             `koanf:"saveLogs"`
	DefaultIconForBack                 string                     `koanf:"defaultIconForBack"`
	AuthOAuth2RedirectURL              string                     `koanf:"authOAuth2RedirectUrl"`
	ListenAddressRestActions           string                     `koanf:"listenAddressRestActions"`
	ListenAddressPrometheus            string                     `koanf:"listenAddressPrometheus"`
	ExternalRestAddress                string                     `koanf:"externalRestAddress"`
	LogLevel                           string                     `koanf:"logLevel"`
	ThemeName                          string                     `koanf:"themeName"`
	ServiceLogs                        ServiceLogsConfig          `koanf:"serviceLogs"`
	ListenAddressSingleHTTPFrontend    string                     `koanf:"listenAddressSingleHTTPFrontend"`
	AuthJwtHmacSecret                  string                     `koanf:"authJwtHmacSecret"`
	AuthJwtCertsURL                    string                     `koanf:"authJwtCertsUrl"`
	DefaultIconForActions              string                     `koanf:"defaultIconForActions"`
	Include                            string                     `koanf:"include"`
	PageTitle                          string                     `koanf:"pageTitle"`
	BannerCSS                          string                     `koanf:"bannerCss"`
	BannerMessage                      string                     `koanf:"bannerMessage"`
	DefaultPopupOnStart                string                     `koanf:"defaultPopupOnStart"`
	ServiceHostMode                    string                     `koanf:"serviceHostMode"`
	DefaultOnClick                     string                     `koanf:"defaultOnClick"`
	AuthJwtCookieName                  string                     `koanf:"authJwtCookieName"`
	AuthJwtHeader                      string                     `koanf:"authJwtHeader"`
	AuthJwtAud                         string                     `koanf:"authJwtAud"`
	ListenAddressWebUI                 string                     `koanf:"listenAddressWebUI"`
	SectionNavigationStyle             string                     `koanf:"sectionNavigationStyle"`
	DefaultIconForDirectories          string                     `koanf:"defaultIconForDirectories"`
	AuthJwtClaimUsername               string                     `koanf:"authJwtClaimUsername"`
	AuthJwtClaimUserGroup              string                     `koanf:"authJwtClaimUserGroup"`
	AuthJwtPubKeyPath                  string                     `koanf:"authJwtPubKeyPath"`
	AuthHttpHeaderUsername             string                     `koanf:"authHttpHeaderUsername"`
	AuthHttpHeaderUserGroup            string                     `koanf:"authHttpHeaderUserGroup"`
	AuthHttpHeaderUserGroupSep         string                     `koanf:"authHttpHeaderUserGroupSep"`
	WebUIDir                           string                     `koanf:"webUIDir"`
	AuthLoginUrl                       string                     `koanf:"authLoginUrl"`
	AuthJwtDomain                      string                     `koanf:"authJwtDomain"`
	Security                           SecurityConfig             `koanf:"security"`
	Actions                            []*Action                  `koanf:"actions"`
	AccessControlLists                 []*AccessControlList       `koanf:"accessControlLists"`
	StyleMods                          []string                   `koanf:"styleMods"`
	AdditionalNavigationLinks          []*NavigationLink          `koanf:"additionalNavigationLinks"`
	Entities                           []*EntityFile              `koanf:"entities"`
	Dashboards                         []*DashboardComponent      `koanf:"dashboards"`
	sourceFiles                        []string
	AuthLocalUsers                     AuthLocalUsersConfig `koanf:"authLocalUsers"`
	LogHistoryPageSize                 int64                `koanf:"logHistoryPageSize"`
	LogDebugOptions                    LogDebugOptions      `koanf:"logDebugOptions"`
	DefaultPermissions                 PermissionsList      `koanf:"defaultPermissions"`
	DefaultPolicy                      ConfigurationPolicy  `koanf:"defaultPolicy"`
	Prometheus                         PrometheusConfig     `koanf:"prometheus"`
	CheckForUpdates                    bool                 `koanf:"checkForUpdates"`
	InsecureAllowDumpJwtClaims         bool                 `koanf:"insecureAllowDumpJwtClaims"`
	InsecureAllowDumpActionMap         bool                 `koanf:"insecureAllowDumpActionMap"`
	InsecureAllowDumpServerDiagnostics bool                 `koanf:"insecureAllowDumpServerDiagnostics"`
	InsecureAllowDumpVars              bool                 `koanf:"insecureAllowDumpVars"`
	InsecureAllowDumpOAuth2UserData    bool                 `koanf:"insecureAllowDumpOAuth2UserData"`
	CronSupportForSeconds              bool                 `koanf:"cronSupportForSeconds"`
	AuthRequireGuestsToLogin           bool                 `koanf:"authRequireGuestsToLogin"`
	EnableCustomJs                     bool                 `koanf:"enableCustomJs"`
	ShowNavigateOnStartIcons           bool                 `koanf:"showNavigateOnStartIcons"`
	ShowNewVersions                    bool                 `koanf:"showNewVersions"`
	ShowNavigation                     bool                 `koanf:"showNavigation"`
	ShowFooter                         bool                 `koanf:"showFooter"`
	UseSingleHTTPFrontend              bool                 `koanf:"useSingleHTTPFrontend"`
	ThemeCacheDisabled                 bool                 `koanf:"themeCacheDisabled"`
	Features                           FeaturesConfig       `koanf:"features"`
}

type AuthLocalUsersConfig struct {
	Users   []*LocalUser `koanf:"users"`
	Enabled bool         `koanf:"enabled"`
}

type LocalUser struct {
	Username  string `koanf:"username"`
	Usergroup string `koanf:"usergroup"`
	Password  string `koanf:"password"`
	ApiKey    string `koanf:"apiKey"`
}

type OAuth2Provider struct {
	AuthUrl            string   `koanf:"authUrl"`
	UserGroupField     string   `koanf:"userGroupField"`
	ClientID           string   `koanf:"clientId"`
	ClientSecret       string   `koanf:"clientSecret"`
	Icon               string   `koanf:"icon"`
	AddToUsergroup     string   `koanf:"addToUsergroup"`
	Title              string   `koanf:"title"`
	WhoamiUrl          string   `koanf:"whoamiUrl"`
	Name               string   `koanf:"name"`
	UsernameField      string   `koanf:"usernameField"`
	TokenUrl           string   `koanf:"tokenUrl"`
	CertBundlePath     string   `koanf:"certBundlePath"`
	Scopes             []string `koanf:"scopes"`
	CallbackTimeout    int      `koanf:"callbackTimeout"`
	InsecureSkipVerify bool     `koanf:"insecureSkipVerify"`
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

type ServiceLogsConfig struct {
	Directory string `koanf:"directory"`
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
	Category     string                `koanf:"category"`
	Type         string                `koanf:"type"`
	Entity       string                `koanf:"entity"`
	Icon         string                `koanf:"icon"`
	CssClass     string                `koanf:"cssClass"`
	Acls         []string              `koanf:"acls"`
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
	config.DefaultOnClick = "nothing"
	config.DefaultPopupOnStart = "nothing"
	config.InsecureAllowDumpVars = false
	config.InsecureAllowDumpServerDiagnostics = false
	config.InsecureAllowDumpActionMap = false
	config.InsecureAllowDumpJwtClaims = false
	config.Prometheus.Enabled = false
	config.Prometheus.DefaultGoMetrics = false
	config.Security.HeaderContentSecurityPolicy = true
	config.Security.ContentSecurityPolicy = ContentSecurityPolicyDefault
	config.Security.HeaderXContentTypeOptions = true
	config.Security.HeaderXFrameOptions = true
	config.Security.XFrameOptions = "DENY"
	config.DefaultIconForActions = "hugeicons:CommandLineIcon"
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
	config.DefaultPolicy.ShowVersionNumber = true

	return &config
}
