package installationinfo

import (
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	"gopkg.in/yaml.v3"
)

var Config *config.Config

type sosReportConfig struct {
	CountOfActions int
	LogLevel       string
}

func configToSosreport(cfg *config.Config) *sosReportConfig {
	return &sosReportConfig{
		CountOfActions: len(cfg.Actions),
		LogLevel:       cfg.LogLevel,
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
