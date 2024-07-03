package installationinfo

import (
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"gopkg.in/yaml.v3"
	"time"
)

var Config *config.Config

type sosReportConfig struct {
	CountOfActions                  int
	CountOfDashboards               int
	LogLevel                        string
	ListenAddressSingleHTTPFrontend string
	ListenAddressWebUI              string
	ListenAddressRestActions        string
	ListenAddressGrpcActions        string
	Timezone                        string
	TimeNow                         string
	ConfigDirectory                 string
	WebuiDirectory                  string
	ThemesDirectory                 string
}

func configToSosreport(cfg *config.Config) *sosReportConfig {
	return &sosReportConfig{
		CountOfActions:                  len(cfg.Actions),
		CountOfDashboards:               len(cfg.Dashboards),
		LogLevel:                        cfg.LogLevel,
		ListenAddressSingleHTTPFrontend: cfg.ListenAddressSingleHTTPFrontend,
		ListenAddressWebUI:              cfg.ListenAddressWebUI,
		ListenAddressRestActions:        cfg.ListenAddressRestActions,
		ListenAddressGrpcActions:        cfg.ListenAddressGrpcActions,
		Timezone:                        time.Now().Location().String(),
		TimeNow:                         time.Now().String(),
		ConfigDirectory:                 cfg.GetDir(),
		WebuiDirectory:                  sv.Get("internal.webuidir"),
		ThemesDirectory:                 sv.Get("internal.themesdir"),
	}
}

func GetSosReport() string {
	ret := ""

	ret += "### SOSREPORT START (copy all text to SOSREPORT END)\n"

	out, _ := yaml.Marshal(Build)
	ret += fmt.Sprintf("# Build: \n%+v\n", string(out))

	out, _ = yaml.Marshal(Runtime)
	ret += fmt.Sprintf("# Runtime:\n%+v\n", string(out))

	out, _ = yaml.Marshal(configToSosreport(Config))
	ret += fmt.Sprintf("# Config:\n%+v\n", string(out))
	ret += "### SOSREPORT END  (copy all text from SOSREPORT START)\n"

	return ret
}
