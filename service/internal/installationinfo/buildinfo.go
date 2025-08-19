package installationinfo

type BuildInfo struct {
	Commit  string
	Version string
	Date    string
}

var Build = &BuildInfo{}
