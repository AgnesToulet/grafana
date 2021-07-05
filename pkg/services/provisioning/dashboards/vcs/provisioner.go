package vcs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/dashboards"
	dashboardprovisioning "github.com/grafana/grafana/pkg/services/provisioning/dashboards"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	ProvisionerName = "gitops-vcs-provisioning"
	PollingPeriod   = 300 // Todo make this configurable
)

type Provisioner struct {
	VCS   vcs.Service        `inject:""`
	Store *sqlstore.SQLStore `inject:""`
	Cfg   *setting.Cfg       `inject:""`
	log   log.Logger
}

// Ensure compliance to the interface
var _ dashboardprovisioning.DashboardProvisioner = &Provisioner{}

func init() {
	registry.Register(&registry.Descriptor{
		Name:         "DashboardVcsProvisioner",
		Instance:     &Provisioner{},
		InitPriority: registry.Low + 1,
	})
}

func (p *Provisioner) Init() error {
	p.log = log.New("dashboards.provisioner")
	return nil
}

func (p *Provisioner) IsDisabled() bool {
	_, gitops := p.Cfg.FeatureToggles["gitops"]

	return !gitops
}

// Provision provisioned all latest datasources files found in VCS should their version have change
func (p *Provisioner) Provision(ctx context.Context) error {
	if p.VCS.IsDisabled() {
		p.log.Warn("cannot provision, VCS service is disabled")
		return nil
	}

	vobjs, err := p.VCS.Latest(ctx, vcs.Dashboard)
	if err != nil {
		p.log.Warn("cannot provision using VCS", err)
		return nil
	}

	dashSvc := dashboards.NewProvisioningService(p.Store)
	provisionedDashboardRefs, err := getProvisionedDashboardsByPath(dashSvc, ProvisionerName)
	if err != nil {
		return err
	}

	for _, obj := range vobjs {
		// Here we assume the dash.Id is correct => dashboard was saved from grafana
		path := getDashboardPath(obj.ID)
		pastProvisioningInfo, alreadyProvisioned := provisionedDashboardRefs[path]

		upToDate := alreadyProvisioned
		if pastProvisioningInfo != nil {
			upToDate = obj.Version == pastProvisioningInfo.CheckSum
		}

		if upToDate {
			continue
		}

		if err = p.saveProvisionedDashboard(dashSvc, obj, pastProvisioningInfo); err != nil {
			return err
		}
	}

	return nil
}

// saveProvisionedDashboard creates the structures to save a dashboard and its provisioning info in store
func (p *Provisioner) saveProvisionedDashboard(dashSvc dashboards.DashboardProvisioningService, obj vcs.VersionedObject, pastProvisioningInfo *models.DashboardProvisioning) error {
	// Unmarshal the dashboard json
	dash := &models.Dashboard{}
	err := json.Unmarshal(obj.Data, dash)
	if err != nil {
		return err
	}
	path := getDashboardPath(obj.ID)

	// Prepare info to save dashboard in store
	dto := dashboards.SaveDashboardDTO{
		OrgId:     dash.OrgId,
		UpdatedAt: dash.Updated,
		Message:   "Provisioned by " + ProvisionerName,
		Overwrite: false,
		Dashboard: dash,
	}

	if dto.Dashboard.Id != 0 {
		dto.Dashboard.Data.Set("id", nil)
		dto.Dashboard.Id = 0
	}

	if pastProvisioningInfo != nil {
		dto.Dashboard.SetId(pastProvisioningInfo.DashboardId)
	}

	// Prepare info to save provisioning information in store
	dp := &models.DashboardProvisioning{
		ExternalId: path,
		Name:       ProvisionerName,
		Updated:    obj.Timestamp,
		CheckSum:   obj.Version,
	}

	// Use the dashboard provisioning service to save the dashboard
	p.log.Debug("saving new dashboard", "provisioner", ProvisionerName,
		"file", path,
		"folderId", dash.FolderId)

	_, err = dashSvc.SaveProvisionedDashboard(&dto, dp)
	if err != nil {
		return err
	}
	return nil
}

// PollChanges calls the Provision method on a ticker
func (p *Provisioner) PollChanges(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Duration(int64(time.Second) * PollingPeriod))
		for {
			select {
			case <-ticker.C:
				p.Provision(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// GetProvisionerResolvedPath returns the path of the config file for the provisioner
// The VCS provisioner only has one path: vcs.Dashboard
func (p *Provisioner) GetProvisionerResolvedPath(name string) string {
	return fmt.Sprintf("%s/%s/", ProvisionerName, vcs.Dashboard)
}

// GetAllowUIUpdatesFromConfig returns whether provisioned dashboards for a given provisioner
// can be updated from frontend or not. Since we handle saving to the VCS storage, it makes
// sense to allow updates.
func (p *Provisioner) GetAllowUIUpdatesFromConfig(name string) bool {
	return true
}

// CleanUpOrphanedDashboards removes dashboards that are not in the VCS storage anymore
func (p *Provisioner) CleanUpOrphanedDashboards() {
	if p.VCS.IsDisabled() {
		p.log.Warn("cannot clean up orphaned dashboards, VCS service is disabled")
		return
	}

	vobjs, err := p.VCS.Latest(context.TODO(), vcs.Dashboard)
	if err != nil {
		p.log.Warn("cannot clean up orphaned dashboards", err)
		return
	}

	dashSvc := dashboards.NewProvisioningService(p.Store)
	provisionedDashboardRefs, err := getProvisionedDashboardsByPath(dashSvc, ProvisionerName)

	for _, obj := range vobjs {
		path := getDashboardPath(obj.ID)

		// Remove all dashboards we have in store from the map
		delete(provisionedDashboardRefs, path)
	}

	// Every remaning dashboard is orphaned
	for _, orphaned := range provisionedDashboardRefs {
		err = sqlstore.DeleteDashboard(&models.DeleteDashboardCommand{Id: orphaned.DashboardId})
	}
}

// getProvisionedDashboardsByPath requests the dashSvc for all provisioning info stored in the database
// and sort them by path (path ex: "dashboard/xUIhNbu0.json")
func getProvisionedDashboardsByPath(dashSvc dashboards.DashboardProvisioningService, provisionerName string) (
	map[string]*models.DashboardProvisioning, error) {
	arr, err := dashSvc.GetProvisionedDashboardData(provisionerName)
	if err != nil {
		return nil, err
	}

	byPath := map[string]*models.DashboardProvisioning{}
	for _, pd := range arr {
		byPath[pd.ExternalId] = pd
	}

	return byPath, nil
}

// getDashboardPath artificially creates a path to identify the data
func getDashboardPath(id string) string {
	return fmt.Sprintf("%s/%s.json", vcs.Dashboard, id)
}
