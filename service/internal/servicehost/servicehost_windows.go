//go:build windows
// +build windows

package servicehost

import (
	log "github.com/sirupsen/logrus"

	"fmt"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"os"
	"path"
	"time"
)

type otWindowsService struct{}

//gocyclo:ignore
func (m *otWindowsService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificExitCode bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	tick := time.Tick(30 * time.Second)

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	log.Info("Service started")

	for {
		select {
		case <-tick:
			log.Info("servicehost Tick")
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Info("Stopping service")
				return false, 0
			case svc.Pause:
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				log.Infof("Unexpected control request: %d", c.Cmd)
			}
		}
	}

	// status <- svc.Status{State: svc.StopPending}
	// return false, 1
}

func setupLogging() {
	logsDir := path.Join(GetConfigFilePath(), "logs")

	os.MkdirAll(logsDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	filename := path.Join(logsDir, fmt.Sprintf("OliveTin-service-%v.log", timestamp))

	log.Infof("Setting up logging to file: %v", filename)

	f, err := os.Create(filename)

	if err != nil {
		log.Infof("Failed to open log file: %v", err)
	} else {
		log.Infof("Switching to log file: %v", f.Name())
		log.SetOutput(f)
		log.Infof("Opened log file: %v", f.Name())
	}
}

func GetConfigFilePath() string {
	programDataDir := path.Join(os.Getenv("ProgramData"), "OliveTin")

	_, err := os.Stat(programDataDir)

	if os.IsNotExist(err) {
		os.MkdirAll(programDataDir, 0755)
	}

	return programDataDir
}

func startServiceHandler(mode string) {
	const serviceName = "OliveTin"

	var err error

	switch mode {
	case "winsvc-debug":
		log.Infof("Running Windows service in debug mode")

		err = debug.Run(serviceName, &otWindowsService{})
	case "winsvc-standard":
		log.Infof("Running Windows service in standard mode")

		err = svc.Run(serviceName, &otWindowsService{})
	case "":
		return
	default:
		log.Fatalf("Unknown servicehost service mode: %s", mode)
	}

	if err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}

	log.Infof("Servicehost handler completed")

	os.Exit(0)

}

func Start(mode string) {
	setupLogging()

	go startServiceHandler(mode)

	// Give some time for the logging to be setup before starting
	// the service to avoid losing any log messages.
	time.Sleep(2 * time.Second)
}
