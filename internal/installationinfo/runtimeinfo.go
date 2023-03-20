package installationinfo

import (
	"bufio"
	"errors"
	"os"
	"runtime"
	"strings"
)

type runtimeInfo struct {
	OS                   string
	OSReleasePrettyName  string
	Arch                 string
	InContainer          bool
	LastBrowserUserAgent string
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

var Runtime = &runtimeInfo{
	OS:                  runtime.GOOS,
	Arch:                runtime.GOARCH,
	InContainer:         isInContainer(),
	OSReleasePrettyName: getOsReleasePrettyName(),
}
