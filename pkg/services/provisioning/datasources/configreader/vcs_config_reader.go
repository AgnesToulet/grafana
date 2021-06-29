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
	configs := []*datasources.Configs{}
	datasourcesMap, err := cr.vcs.Latest(ctx, vcs.Datasource)
	if err != nil {
		return nil, err
	}

	for _, versionedDc := range datasourcesMap {
		cfg := datasources.ConfigsV1{Log: cr.log}
		cfg.APIVersion = 1

		dcCfg := datasources.UpsertDataSourceFromConfigV1{}
		err := json.Unmarshal(versionedDc.Data, &dcCfg)
		if err != nil {
			return nil, err
		}

		cfg.Datasources = append(cfg.Datasources, &dcCfg)
		configs = append(configs, cfg.MapToDatasourceFromConfig(1))
	}

	return configs, nil
}
