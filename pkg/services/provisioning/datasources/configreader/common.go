package configreader

import (
	"fmt"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/provisioning/utils"
)

func validateDefaultUniqueness(logger log.Logger, datasourcesCfg []*datasources.Configs) error {
	defaultCount := map[int64]int{}
	for i := range datasourcesCfg {
		if datasourcesCfg[i].Datasources == nil {
			continue
		}

		for _, ds := range datasourcesCfg[i].Datasources {
			if ds.OrgID == 0 {
				ds.OrgID = 1
			}

			if err := validateAccessAndOrgID(logger, ds); err != nil {
				return fmt.Errorf("failed to provision %q data source: %w", ds.Name, err)
			}

			if ds.IsDefault {
				defaultCount[ds.OrgID]++
				if defaultCount[ds.OrgID] > 1 {
					return datasources.ErrInvalidConfigToManyDefault
				}
			}
		}

		for _, ds := range datasourcesCfg[i].DeleteDatasources {
			if ds.OrgID == 0 {
				ds.OrgID = 1
			}
		}
	}

	return nil
}

func validateAccessAndOrgID(logger log.Logger, ds *datasources.UpsertDataSourceFromConfig) error {
	if err := utils.CheckOrgExists(ds.OrgID); err != nil {
		return err
	}

	if ds.Access == "" {
		ds.Access = models.DS_ACCESS_PROXY
	}

	if ds.Access != models.DS_ACCESS_DIRECT && ds.Access != models.DS_ACCESS_PROXY {
		logger.Warn("invalid access value, will use 'proxy' instead", "value", ds.Access)
		ds.Access = models.DS_ACCESS_PROXY
	}
	return nil
}
