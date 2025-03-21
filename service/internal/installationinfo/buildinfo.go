package installationinfo

type buildInfo struct {
	Commit  string
	Version string
	Date    string
}

var Build = &buildInfo{}
