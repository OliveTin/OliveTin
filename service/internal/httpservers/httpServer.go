package httpservers

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	config "github.com/OliveTin/OliveTin/internal/config"
)

type OliveTinServer struct {
	cfg                  *config.Config
	WebUiServer          *http.Server
	PrometheusServer     *http.Server
	RestAPIServer        *http.Server
	SingleFrontendServer *http.Server
	done                 chan struct{}
}

func CreateOliveTinServers(cfg *config.Config) *OliveTinServer {
	return &OliveTinServer{
		cfg:  cfg,
		done: make(chan struct{}),
	}
}

// StartServers will start 3 HTTP servers. The WebUI, the Rest API, and a proxy
// for both of them.
func (s *OliveTinServer) StartServers() {
	s.RestAPIServer = startRestAPIServer(s.cfg)
	s.WebUiServer = startWebUIServer(s.cfg)

	go func() {
		err := s.WebUiServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		err := s.RestAPIServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if s.cfg.UseSingleHTTPFrontend {
		s.SingleFrontendServer = StartSingleHTTPFrontend(s.cfg)

		go func() {
			err := s.SingleFrontendServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}

	if s.cfg.Prometheus.Enabled {
		s.PrometheusServer = StartPrometheus(s.cfg)
		go func() {
			err := s.PrometheusServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}

	for {
		select {
		case <-s.done:
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (s *OliveTinServer) Stop() {
	if s.WebUiServer != nil {
		s.WebUiServer.Shutdown(context.Background())
		s.WebUiServer = nil
	}
	if s.SingleFrontendServer != nil {
		s.SingleFrontendServer.Shutdown(context.Background())
		s.SingleFrontendServer = nil
	}
	if s.PrometheusServer != nil {
		s.PrometheusServer.Shutdown(context.Background())
		s.PrometheusServer = nil
	}
	if s.RestAPIServer != nil {
		s.RestAPIServer.Shutdown(context.Background())
		s.RestAPIServer = nil
	}
	close(s.done)
}
