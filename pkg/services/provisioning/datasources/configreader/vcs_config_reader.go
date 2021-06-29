package configreader

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/vcs"
)

type vcsConfigReader struct {
	log log.Logger
	vcs vcs.Service
}

func NewVCSConfigReader(log log.Logger, vcs vcs.Service) datasources.ConfigReader {
	return &vcsConfigReader{log: log, vcs: vcs}
}

func (cr *vcsConfigReader) ReadConfigs(ctx context.Context) ([]*datasources.Configs, error) {
	// Get all versioned datasources
	datasourcesMap, err := cr.vcs.Latest(ctx, vcs.Datasource)
	if err != nil {
		return nil, err
	}
	if len(datasourcesMap) == 0 {
		return []*datasources.Configs{}, nil
	}

	// Prepare a config for provisioning
	cfg := datasources.Configs{
		APIVersion:        1, // TODO manage to get this somehow?
		Datasources:       []*datasources.UpsertDataSourceFromConfig{},
		DeleteDatasources: nil,
	}
	configs := []*datasources.Configs{&cfg}

	// Parse all versioned objects
	for _, versionedDatasource := range datasourcesMap {
		ds, err := cr.parseDatasourceConfig(cfg.APIVersion, versionedDatasource)
		if err != nil {
			return nil, err
		}
		cfg.Datasources = append(cfg.Datasources, ds)
	}

	// Validate configuration
	err = validateDefaultUniqueness(cr.log, configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (cr *vcsConfigReader) parseDatasourceConfig(apiVersion int64, obj vcs.VersionedObject) (*datasources.UpsertDataSourceFromConfig, error) {
	dcCfg := upsertDataSourceFromConfigV1{}

	err := json.Unmarshal(obj.Data, &dcCfg)
	if err != nil {
		return nil, err
	}

	return dcCfg.mapToUpsertDataSourceFromConfig(), nil
}
