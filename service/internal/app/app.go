package app

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entityfiles"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/OliveTin/OliveTin/internal/grpcapi"
	"github.com/OliveTin/OliveTin/internal/httpservers"
	"github.com/OliveTin/OliveTin/internal/oncalendarfile"
	"github.com/OliveTin/OliveTin/internal/oncron"
	"github.com/OliveTin/OliveTin/internal/onfileindir"
	"github.com/OliveTin/OliveTin/internal/onstartup"
	"github.com/OliveTin/OliveTin/internal/updatecheck"
	"github.com/OliveTin/OliveTin/internal/websocket"
)

type olivetin struct {
	cfg  *config.Config
	ots  *httpservers.OliveTinServer
	grpc *grpc.Server
	done chan struct{}
}

func CreateOliveTin(cfg *config.Config) *olivetin {
	log.WithFields(log.Fields{
		"configDir": cfg.GetDir(),
	}).Infof("OliveTin started")

	log.Debugf("Config: %+v", cfg)

	executor := executor.DefaultExecutor(cfg)
	executor.RebuildActionMap()
	executor.AddListener(websocket.ExecutionListener)
	config.AddListener(executor.RebuildActionMap)

	go onstartup.Execute(cfg, executor)
	go oncron.Schedule(cfg, executor)
	go onfileindir.WatchFilesInDirectory(cfg, executor)
	go oncalendarfile.Schedule(cfg, executor)

	entityfiles.AddListener(websocket.OnEntityChanged)
	entityfiles.AddListener(executor.RebuildActionMap)
	go entityfiles.SetupEntityFileWatchers(cfg)

	go updatecheck.StartUpdateChecker(cfg)

	srv := grpcapi.Start(cfg, executor)

	ots := httpservers.CreateOliveTinServers(cfg)

	return &olivetin{
		cfg:  cfg,
		ots:  ots,
		grpc: srv,
		done: make(chan struct{}),
	}
}

func (o *olivetin) Start() {

	go func() {
		lis, err := net.Listen("tcp", o.cfg.ListenAddressGrpcActions)
		if err != nil {
			log.Fatalf("Failed to listen - %v", err)
		}
		err = o.grpc.Serve(lis)
		if err != nil {
			log.Fatalf("Could not start gRPC Server - %v", err)
		}
	}()

	go o.ots.StartServers()
	for {
		select {
		case <-o.done:
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}
func (o *olivetin) Stop() {
	o.grpc.Stop()
	o.ots.Stop()
	close(o.done)
}
