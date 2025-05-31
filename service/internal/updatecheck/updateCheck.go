package updatecheck

import (
	"encoding/json"
	"github.com/Masterminds/semver"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/installationinfo"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

type versionMapType struct {
	ApiVersion int
	Latest     string
	History    map[string]string
}

// StartUpdateChecker will start a job that runs periodically, checking
// for updates.
func StartUpdateChecker(cfg *config.Config) {
	if !cfg.CheckForUpdates {
		installationinfo.Runtime.AvailableVersion = "none"
		log.Infof("Update checking is disabled")
		return
	}

	s := cron.New()

	// Several values have been tried here.
	// 1st: Every 24h - very spammy.
	// 2nd: Every 7d - (168 hours - much more reasonable, but it checks in at the same time/day each week.
	// Current: Every 100h is not so spammy, and has the advantage that the checkin time "shifts" hours.
	s.AddFunc("@every 100h", func() {
		actualCheckForUpdate()
	})

	go actualCheckForUpdate() // On startup

	go s.Start()
}

func parseVersion(input []byte) string {
	versionMap := &versionMapType{}

	err := json.Unmarshal(input, &versionMap)

	if err != nil {
		log.Errorf("Update check unmarshal failure: %v", err)
		return "none"
	} else {
		log.Infof("Update check remote version: %+v, latest version: %+v", versionMap.Latest, installationinfo.Build.Version)

		if installationinfo.Build.Version == versionMap.Latest {
			return "none"
		} else {
			return parseIfVersionIsLater(installationinfo.Build.Version, versionMap.Latest)
		}
	}
}

func parseIfVersionIsLater(currentString string, latestString string) string {
	currentVersion, errCurrent := semver.NewVersion(currentString)
	latestVersion, errLatest := semver.NewVersion(latestString)

	if errCurrent != nil || errLatest != nil {
		log.Warnf("Version parse failure: %v %v", errCurrent, errLatest)

		return "version-parse-failure"
	}

	if latestVersion.GreaterThan(currentVersion) {
		return latestString
	}

	return "none"
}

func doRequest() string {
	req, err := http.NewRequest("GET", "http://update-check.olivetin.app/versions.json", nil)

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return "none"
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return "none"
	}

	versionMap, _ := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	return parseVersion(versionMap)
}

func actualCheckForUpdate() {
	if installationinfo.Build.Version == "dev" && os.Getenv("OLIVETIN_FORCE_UPDATE_CHECK") == "" {
		installationinfo.Runtime.AvailableVersion = "you-are-using-a-dev-build"
	} else {
		installationinfo.Runtime.AvailableVersion = doRequest()
	}

	log.WithFields(log.Fields{
		"CurrentVersion": installationinfo.Build.Version,
		"NewVersion":     installationinfo.Runtime.AvailableVersion,
	}).Infof("Update check complete")
}
