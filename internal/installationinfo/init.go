package installationinfo

import (
	"fmt"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

func init() {
	sv.Set("OliveTin.build.commit", Build.Commit)
	sv.Set("OliveTin.build.version", Build.Version)
	sv.Set("OliveTin.build.date", Build.Date)
	sv.Set("OliveTin.runtime.os", Runtime.OS)
	sv.Set("OliveTin.runtime.os.pretty", Runtime.OSReleasePrettyName)
	sv.Set("OliveTin.runtime.arch", Runtime.Arch)
	sv.Set("OliveTin.runtime.incontainer", fmt.Sprintf("%v", Runtime.InContainer))
}
