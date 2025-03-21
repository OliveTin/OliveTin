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
	SshFoundKey          string
	SshFoundConfig       string
	AvailableVersion     string
}

var Runtime = &runtimeInfo{
	OS:                  runtime.GOOS,
	Arch:                runtime.GOARCH,
	InContainer:         isInContainer(),
	OSReleasePrettyName: getOsReleasePrettyName(),
	User:                os.Getenv("USER"),
	Uid:                 os.Getenv("UID"),
	SshFoundKey:         searchForSshKey(),
	SshFoundConfig:      searchForSshConfig(),
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

func searchForSshKey() string {
	if fileExists("/config/ssh/id_rsa") {
		return "/config/ssh/id_rsa"
	}

	return searchForHomeFile(".ssh/id_rsa")
}

func searchForSshConfig() string {
	if fileExists("/config/ssh/config") {
		return "/config/ssh/config"
	}

	return searchForHomeFile(".ssh/config")
}

func searchForHomeFile(file string) string {
	path, _ := filepath.Abs(path.Join(os.Getenv("HOME"), file))

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "not found at " + path
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
