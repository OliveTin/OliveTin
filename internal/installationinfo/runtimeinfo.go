package installationinfo

import (
	"bufio"
	"errors"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type runtimeInfo struct {
	OS                   string
	OSReleasePrettyName  string
	Arch                 string
	InContainer          bool
	LastBrowserUserAgent string
	User                 string
	Uid                  string
	FoundSshKey          string
	AvailableVersion     string
}

var Runtime = &runtimeInfo{
	OS:                  runtime.GOOS,
	Arch:                runtime.GOARCH,
	InContainer:         isInContainer(),
	OSReleasePrettyName: getOsReleasePrettyName(),
	User:                os.Getenv("USER"),
	Uid:                 os.Getenv("UID"),
}

func refreshRuntimeInfo() {
	Runtime.FoundSshKey = searchForSshKey()
}

func searchForSshKey() string {
	path, _ := filepath.Abs(path.Join(os.Getenv("HOME"), ".ssh/id_rsa"))

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "none-found at " + path
}

func isInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func getOsReleasePrettyName() string {
	handle, err := os.Open("/etc/os-release")

	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(handle)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "PRETTY_NAME") {
			return line
		}
	}

	handle.Close()

	return "notfound"
}
