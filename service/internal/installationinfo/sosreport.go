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

type sosReportConfig struct {
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

func configToSosreport(cfg *config.Config) *sosReportConfig {
	return &sosReportConfig{
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

func GetSosReport(redactVersion bool) string {
	ret := ""

	ret += "### SOSREPORT START (copy all text to SOSREPORT END)\n"

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

	out, _ = yaml.Marshal(configToSosreport(Config))
	ret += fmt.Sprintf("# Config:\n%+v\n", string(out))
	ret += "### SOSREPORT END  (copy all text from SOSREPORT START)\n"

	return ret
}
