package datasources

import (
	"errors"
	"path/filepath"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/setting"

	"github.com/grafana/grafana/pkg/infra/log"

	"github.com/grafana/grafana/pkg/models"
)

var (
	// ErrInvalidConfigToManyDefault indicates that multiple datasource in the provisioning files
	// contains more than one datasource marked as default.
	ErrInvalidConfigToManyDefault = errors.New("datasource.yaml config is invalid. Only one datasource per organization can be marked as default")
)

// Provision scans a directory for provisioning config files
// and provisions the datasource in those files.
func (dc *DatasourceProvisioner) Provision() error {
	configPath := filepath.Join(dc.cfg.ProvisioningPath, "datasources")
	return dc.applyChanges(configPath)
}

// DatasourceProvisioner is responsible for provisioning datasources based on
// configuration read by the `configReader`
type DatasourceProvisioner struct {
	log         log.Logger
	cfg         *setting.Cfg
	cfgProvider *configReader
}

func NewDatasourceProvisioner(cfg *setting.Cfg) DatasourceProvisioner {
	logger := log.New("accesscontrol.provisioner")
	return DatasourceProvisioner{
		log:         logger,
		cfg:         cfg,
		cfgProvider: &configReader{log: logger},
	}
}

func (dc *DatasourceProvisioner) apply(cfg *configs) error {
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

func (dc *DatasourceProvisioner) applyChanges(configPath string) error {
	configs, err := dc.cfgProvider.readConfig(configPath)
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

func (dc *DatasourceProvisioner) deleteDatasources(dsToDelete []*deleteDatasourceConfig) error {
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
