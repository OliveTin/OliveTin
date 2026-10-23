package api

import (
	"sync"
	"testing"

	"github.com/OliveTin/OliveTin/internal/executor"
)

// TestBuildDefaultDashboardConcurrentWithRebuild guards against a crash when the
// default dashboard is built (e.g. by Init) while a config or entity reload
// rebuilds the action map. Without the read lock in buildDefaultDashboard, the
// Go runtime aborts with "concurrent map iteration and map write"; run with
// -race to see the data race directly.
func TestBuildDefaultDashboardConcurrentWithRebuild(t *testing.T) {
	cfg, _, adminUser := buildViewPermissionTestConfig(t)
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	rr := &DashboardRenderRequest{
		AuthenticatedUser: adminUser,
		cfg:               cfg,
		ex:                ex,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for range 200 {
			ex.RebuildActionMap()
		}
	}()

	go func() {
		defer wg.Done()
		for range 200 {
			buildDefaultDashboard(rr)
		}
	}()

	wg.Wait()
}
