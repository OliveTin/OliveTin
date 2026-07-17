package installationinfo

import (
	"fmt"
	"time"

	config "github.com/OliveTin/OliveTin/internal/config"
	"gopkg.in/yaml.v3"
)

var (
	Config *config.Config
)

type serverDiagnosticsConfig struct {
	CountOfActions                  int
	CountOfDashboards               int
	LogLevel                        string
	ListenAddressSingleHTTPFrontend string
	ListenAddressWebUI              string
	ListenAddressRestActions        string
	Timezone                        string
	TimeNow                         string
	ConfigDirectory                 string
	WebuiDirectory                  string
}

func configToServerDiagnostics(cfg *config.Config) *serverDiagnosticsConfig {
	return &serverDiagnosticsConfig{
		CountOfActions:                  len(cfg.Actions),
		CountOfDashboards:               len(cfg.Dashboards),
		LogLevel:                        cfg.LogLevel,
		ListenAddressSingleHTTPFrontend: cfg.ListenAddressSingleHTTPFrontend,
		ListenAddressWebUI:              cfg.ListenAddressWebUI,
		ListenAddressRestActions:        cfg.ListenAddressRestActions,
		Timezone:                        time.Now().Location().String(),
		TimeNow:                         time.Now().String(),
		ConfigDirectory:                 cfg.GetDir(),
		WebuiDirectory:                  cfg.WebUIDir,
	}
}

func GetServerDiagnostics(redactVersion bool) string {
	ret := ""

	ret += "### SERVER DIAGNOSTICS START (copy all text to SERVER DIAGNOSTICS END)\n"

	buildForReport := *Build
	if redactVersion {
		buildForReport.Version = "[redacted]"
	}
	out, _ := yaml.Marshal(&buildForReport)
	ret += fmt.Sprintf("# Build: \n%+v\n", string(out))

	runtimeForReport := *Runtime
	if redactVersion {
		runtimeForReport.AvailableVersion = "[redacted]"
	}
	out, _ = yaml.Marshal(&runtimeForReport)
	ret += fmt.Sprintf("# Runtime:\n%+v\n", string(out))

	out, _ = yaml.Marshal(configToServerDiagnostics(Config))
	ret += fmt.Sprintf("# Config:\n%+v\n", string(out))
	ret += "### SERVER DIAGNOSTICS END  (copy all text from SERVER DIAGNOSTICS START)\n"

	return ret
}
