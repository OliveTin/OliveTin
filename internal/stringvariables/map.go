/**
 * The ephemeralvariablemap is used "only" for variable substitution in config
 * titles, shell arguments, etc, in the foorm of {{ key }}, like Jinja2.
 *
 * OliveTin itself really only ever "writes" to this map, mostly by loading
 * EntityFiles, and the only form of "reading" is for the variable substitution
 * in configs.
 */

package stringvariables

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	contents map[string]string

	metricSvCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "olivetin_sv_count",
		Help: "The number entries in the sv map",
	})
)

func init() {
	contents = make(map[string]string)
}

func Get(key string) string {
	v, ok := contents[key]

	if !ok {
		return ""
	} else {
		return v
	}
}

func GetAll() map[string]string {
	return contents
}

func Set(key string, value string) {
	contents[key] = value

	metricSvCount.Set(float64(len(contents)))
}
