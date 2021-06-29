package provisioner

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources/configreader"

	"github.com/grafana/grafana/pkg/setting"

	"github.com/grafana/grafana/pkg/infra/log"

	"github.com/grafana/grafana/pkg/models"
)

// Provision scans a directory for provisioning config files
// and provisions the datasource in those files.
func (dc *DatasourceProvisioner) Provision(ctx context.Context) error {
	return dc.applyChanges(ctx)
}

// DatasourceProvisioner is responsible for provisioning datasources based on
// configuration read by the `configReader`
type DatasourceProvisioner struct {
	log         log.Logger
	cfg         *setting.Cfg
	cfgProvider datasources.ConfigReader
}

func NewDatasourceProvisioner(cfg *setting.Cfg) DatasourceProvisioner {
	logger := log.New("accesscontrol.provisioner")
	configPath := filepath.Join(cfg.ProvisioningPath, "datasources")
	return DatasourceProvisioner{
		log:         logger,
		cfg:         cfg,
		cfgProvider: configreader.NewDiskConfigReader(logger, configPath),
	}
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
			insertCmd := datasources.CreateInsertCommand(ds)
			if err := bus.Dispatch(insertCmd); err != nil {
				return err
			}
		} else {
			dc.log.Debug("updating datasource from configuration", "name", ds.Name, "uid", ds.UID)
			updateCmd := datasources.CreateUpdateCommand(ds, cmd.Result.Id)
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
