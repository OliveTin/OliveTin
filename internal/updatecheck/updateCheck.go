package updatecheck

import (
	config "github.com/jamesread/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron"
	"runtime"
	"net/http"
	"encoding/json"
	"bytes"
	"time"
	"io/ioutil"
)

type UpdateRequest struct {
	CurrentVersion string
	CurrentCommit string
	OS string
	Arch string
	MachineId string
}

func machineId() string {
	v, err := machineid.ProtectedID("OliveTin")

	if err != nil {
		log.Warnf("Error getting machine ID: %v", err)
		return "?"
	}

	return v;
}

func CheckForUpdate(currentVersion string, currentCommit string, cfg *config.Config) {
	payload := UpdateRequest {
		CurrentVersion: currentVersion,
		CurrentCommit: currentCommit,
		OS: runtime.GOOS,
		Arch: runtime.GOARCH,
		MachineId: machineId(),
	}

	s := gocron.NewScheduler(time.UTC)

	s.Every(7).Days().Do(func() {
		actualCheckForUpdate(payload)
	})

	s.StartAsync()
}

func actualCheckForUpdate(payload UpdateRequest) {
	jsonUpdateRequest, err := json.Marshal(payload)

	req, err := http.NewRequest("POST", "http://update-check.olivetin.app", bytes.NewReader(jsonUpdateRequest))

	if err != nil {
		log.Errorf("Update check failed %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Errorf("Update check failed %v", err)
	} else {
		newVersion, _ := ioutil.ReadAll(resp.Body)

		log.WithFields(log.Fields {
			"NewVersion": string(newVersion),
		}).Infof("Update check complete");

		defer resp.Body.Close();
	}

}
