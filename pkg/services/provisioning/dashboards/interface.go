package dashboards

import (
	"context"

	"github.com/grafana/grafana/pkg/dashboards"
)

// DashboardProvisioner is responsible for syncing dashboard from disk to
// Grafana's database.
type DashboardProvisioner interface {
	Provision() error
	PollChanges(ctx context.Context)
	GetProvisionerResolvedPath(name string) string
	GetAllowUIUpdatesFromConfig(name string) bool
	CleanUpOrphanedDashboards()
}

// DashboardProvisionerFactory creates DashboardProvisioners based on input
type DashboardProvisionerFactory func(string, dashboards.Store) (DashboardProvisioner, error)
