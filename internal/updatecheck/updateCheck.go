package updatecheck

import (
	"bytes"
	"encoding/json"
	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron"
	config "github.com/jamesread/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"
)

type updateRequest struct {
	CurrentVersion string
	CurrentCommit  string
	OS             string
	Arch           string
	MachineID      string
}

var AvailableVersion = "none"
var CurrentVersion = "?"

func machineID() string {
	v, err := machineid.ProtectedID("OliveTin")

	if err != nil {
		log.Warnf("Error getting machine ID: %v", err)
		return "?"
	}

	return v
}

// StartUpdateChecker will start a job that runs periodically, checking
// for updates.
func StartUpdateChecker(currentVersion string, currentCommit string, cfg *config.Config) {
	if !cfg.CheckForUpdates {
		log.Warn("Update checking is disabled")
		return
	}

	CurrentVersion = currentVersion

	payload := updateRequest{
		CurrentVersion: currentVersion,
		CurrentCommit:  currentCommit,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		MachineID:      machineID(),
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

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return
	}

	AvailableVersion = doRequest(jsonUpdateRequest)

	log.WithFields(log.Fields{
		"NewVersion": AvailableVersion,
	}).Infof("Update check complete")
}
