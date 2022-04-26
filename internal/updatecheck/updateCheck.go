package updatecheck

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"
)

type updateRequest struct {
	CurrentVersion string
	CurrentCommit  string
	OS             string
	Arch           string
	InstallationID string
	InContainer    bool
}

// AvailableVersion is updated when checking with the update service.
var AvailableVersion = "none"

// CurrentVersion is set by the main cmd (which is in tern set as a compile constant)
var CurrentVersion = "?"

func installationID(filename string) string {
	content := "unset"
	contentBytes, err := ioutil.ReadFile(filename)

	if err != nil {
		fileHandle, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

		if err != nil {
			log.Warnf("Could not read + create installation ID file: %v", err)
			return "cant-create"
		}

		content = uuid.NewString()
		fileHandle.WriteString(content)
		fileHandle.Close()
	} else {
		content = string(contentBytes)

		_, err := uuid.Parse(content)

		if err != nil {
			log.Errorf("Invalid installation ID, %v", err)
			content = "invalid-installation-id"
		}
	}

	log.WithFields(log.Fields{
		"content": content,
		"from":    filename,
	}).Infof("Installation ID")

	return content
}

func isInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// StartUpdateChecker will start a job that runs periodically, checking
// for updates.
func StartUpdateChecker(currentVersion string, currentCommit string, cfg *config.Config, configDir string) {
	CurrentVersion = currentVersion

	if !cfg.CheckForUpdates {
		log.Warn("Update checking is disabled")
		return
	}

	payload := updateRequest{
		CurrentVersion: currentVersion,
		CurrentCommit:  currentCommit,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		InstallationID: installationID(configDir + "/installation-id.txt"),
		InContainer:    isInContainer(),
	}

	s := gocron.NewScheduler(time.UTC)

	s.Every(7).Days().Do(func() {
		actualCheckForUpdate(payload)
	})

	s.StartAsync()
}

func doRequest(jsonUpdateRequest []byte) string {
	req, err := http.NewRequest("POST", "http://update-check.olivetin.app", bytes.NewBuffer(jsonUpdateRequest))

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return ""
	}

	newVersion, _ := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	return string(newVersion)
}

func actualCheckForUpdate(payload updateRequest) {
	jsonUpdateRequest, err := json.Marshal(payload)

	log.Debugf("Update request payload: %+v", payload)

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return
	}

	AvailableVersion = doRequest(jsonUpdateRequest)

	log.WithFields(log.Fields{
		"NewVersion": AvailableVersion,
	}).Infof("Update check complete")
}
