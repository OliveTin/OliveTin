/**
 * The ephemeralvariablemap is used "only" for variable substitution in config
 * titles, shell arguments, etc, in the foorm of {{ key }}, like Jinja2.
 *
 * OliveTin itself really only ever "writes" to this map, mostly by loading
 * EntityFiles, and the only form of "reading" is for the variable substitution
 * in configs.
 */

package stringvariables

var Contents map[string]string

func init() {
	Contents = make(map[string]string)
}

func Get(key string) string {
	v, ok := Contents[key]

	if !ok {
		return ""
	} else {
		return v
	}
}
