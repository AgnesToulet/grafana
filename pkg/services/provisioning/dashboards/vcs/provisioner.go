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

func (p *Provisioner) Provision() error {
	if p.VCS.IsDisabled() {
		p.log.Warn("cannot provision, VCS service is disabled")
		return nil
	}

	vobjs, err := p.VCS.Latest(context.TODO(), vcs.Dashboard)
	if err != nil {
		p.log.Warn("cannot provision using VCS", err)
		return nil
	}

	dashSvc := dashboards.NewProvisioningService(p.Store)
	provisionedDashboardRefs, err := getProvisionedDashboardsByPath(dashSvc, ProvisionerName)

	for _, obj := range vobjs {
		dash := &models.Dashboard{}
		err := json.Unmarshal(obj.Data, dash)
		if err != nil {
			return err
		}
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

		p.log.Debug("saving new dashboard", "provisioner", ProvisionerName,
			"file", path,
			"folderId", dash.FolderId)

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

		if alreadyProvisioned {
			dto.Dashboard.SetId(pastProvisioningInfo.DashboardId)
		}

		dp := &models.DashboardProvisioning{
			ExternalId: path,
			Name:       ProvisionerName,
			Updated:    obj.Timestamp,
			CheckSum:   obj.Version,
		}

		_, err = dashSvc.SaveProvisionedDashboard(&dto, dp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provisioner) PollChanges(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(int64(time.Second) * PollingPeriod))
	for {
		select {
		case <-ticker.C:
			p.Provision()
		case <-ctx.Done():
			return
		}
	}
}

func (p *Provisioner) GetProvisionerResolvedPath(name string) string {
	return string(vcs.Dashboard)
}

func (p *Provisioner) GetAllowUIUpdatesFromConfig(name string) bool {
	return true
}

func (p *Provisioner) CleanUpOrphanedDashboards() {

}

func getProvisionedDashboardsByPath(service dashboards.DashboardProvisioningService, provisionerName string) (
	map[string]*models.DashboardProvisioning, error) {
	arr, err := service.GetProvisionedDashboardData(provisionerName)
	if err != nil {
		return nil, err
	}

	byPath := map[string]*models.DashboardProvisioning{}
	for _, pd := range arr {
		byPath[pd.ExternalId] = pd
	}

	return byPath, nil
}

func getDashboardPath(id string) string {
	return fmt.Sprintf("%s/%s.json", vcs.Dashboard, id)
}
