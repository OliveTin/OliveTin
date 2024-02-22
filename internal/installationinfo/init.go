package installationinfo

import (
	"fmt"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

func init() {
	sv.Contents["OliveTin.build.commit"] = Build.Commit
	sv.Contents["OliveTin.build.version"] = Build.Version
	sv.Contents["OliveTin.build.date"] = Build.Date
	sv.Contents["OliveTin.runtime.os"] = Runtime.OS
	sv.Contents["OliveTin.runtime.os.pretty"] = Runtime.OSReleasePrettyName
	sv.Contents["OliveTin.runtime.arch"] = Runtime.Arch
	sv.Contents["OliveTin.runtime.incontainer"] = fmt.Sprintf("%v", Runtime.InContainer)
}
