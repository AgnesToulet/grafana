package provisioner

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources/configreader"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/setting"
)

// TODO find a way to have the ProvisioningService priority without cyclic import
func init() {
	// ProvisioningService priority is Low and since Provision runs upon its
	// start, we need this service to be initialized before.
	registry.Register(&registry.Descriptor{
		Name:         "DatasourceProvisioner",
		Instance:     &DatasourceProvisioner{},
		InitPriority: registry.Low + 1,
	})
}

// DatasourceProvisioner is responsible for provisioning datasources based on
// configuration read by the `configReader`
type DatasourceProvisioner struct {
	Cfg         *setting.Cfg `inject:""`
	VCS         vcs.Service  `inject:""`
	log         log.Logger
	cfgProvider datasources.ConfigReader
}

func (dc *DatasourceProvisioner) Init() error {
	var configReader datasources.ConfigReader

	dc.log = log.New("datasources.provisioner")

	// Use feature toggle to read configs from files or Version Control System
	if gitops, ok := dc.Cfg.FeatureToggles["gitops"]; ok && gitops {
		configReader = configreader.NewVCSConfigReader(dc.log, dc.VCS)
	} else {
		configPath := filepath.Join(dc.Cfg.ProvisioningPath, "datasources")
		configReader = configreader.NewDiskConfigReader(dc.log, configPath)
	}

	dc.cfgProvider = configReader
	return nil
}

// Provision scans a directory for provisioning config files
// and provisions the datasource in those files.
func (dc *DatasourceProvisioner) Provision(ctx context.Context) error {
	return dc.applyChanges(ctx)
}

func (dc *DatasourceProvisioner) apply(cfg *datasources.Configs) error {
	if err := dc.deleteDatasources(cfg.DeleteDatasources); err != nil {
		return err
	}

	for _, ds := range cfg.Datasources {
		cmd := &models.GetDataSourceQuery{OrgId: ds.OrgID, Name: ds.Name}
		err := bus.Dispatch(cmd)
		if err != nil && !errors.Is(err, models.ErrDataSourceNotFound) {
			return err
		}

		if errors.Is(err, models.ErrDataSourceNotFound) {
			dc.log.Info("inserting datasource from configuration ", "name", ds.Name, "uid", ds.UID)
			insertCmd := createInsertCommand(ds)
			if err := bus.Dispatch(insertCmd); err != nil {
				return err
			}
		} else {
			dc.log.Debug("updating datasource from configuration", "name", ds.Name, "uid", ds.UID)
			updateCmd := createUpdateCommand(ds, cmd.Result.Id)
			if err := bus.Dispatch(updateCmd); err != nil {
				return err
			}
		}
	}

	return nil
}

func (dc *DatasourceProvisioner) applyChanges(ctx context.Context) error {
	configs, err := dc.cfgProvider.ReadConfigs(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := dc.apply(cfg); err != nil {
			return err
		}
	}

	return nil
}

func (dc *DatasourceProvisioner) deleteDatasources(dsToDelete []*datasources.DeleteDatasourceConfig) error {
	for _, ds := range dsToDelete {
		cmd := &models.DeleteDataSourceCommand{OrgID: ds.OrgID, Name: ds.Name}
		if err := bus.Dispatch(cmd); err != nil {
			return err
		}

		if cmd.DeletedDatasourcesCount > 0 {
			dc.log.Info("deleted datasource based on configuration", "name", ds.Name)
		}
	}

	return nil
}
