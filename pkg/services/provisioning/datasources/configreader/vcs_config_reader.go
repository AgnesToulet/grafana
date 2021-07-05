package configreader

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
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
		cr.log.Warn("cannot provision using VCS", err)
		return nil, nil
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
		ds, err := parseDatasource(cfg.APIVersion, versionedDatasource)
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

func parseDatasourceJSONData(jsonData *simplejson.Json) (map[string]interface{}, error) {
	var err error = nil
	var res map[string]interface{}
	if jsonData != nil {
		res, err = jsonData.Map()
	}
	return res, err
}

func parseDatasource(apiVersion int64, obj vcs.VersionedObject) (*datasources.UpsertDataSourceFromConfig, error) {
	ds := models.DataSource{}

	err := json.Unmarshal(obj.Data, &ds)
	if err != nil {
		return nil, err
	}

	jsonData, err := parseDatasourceJSONData(ds.JsonData)
	if err != nil {
		return nil, fmt.Errorf("manage to unmarshal datasource but could not unmarshal jsonData: %w", err)
	}

	dsCfg := datasources.UpsertDataSourceFromConfig{
		OrgID:             ds.OrgId,
		Version:           ds.Version,
		Name:              ds.Name,
		Type:              ds.Type,
		Access:            string(ds.Access),
		URL:               ds.Url,
		Password:          ds.Password,
		User:              ds.User,
		Database:          ds.Database,
		BasicAuth:         ds.BasicAuth,
		BasicAuthUser:     ds.BasicAuthUser,
		BasicAuthPassword: ds.BasicAuthPassword,
		WithCredentials:   ds.WithCredentials,
		IsDefault:         ds.IsDefault,
		JSONData:          jsonData,
		SecureJSONData:    ds.SecureJsonData.Decrypt(),
		Editable:          !ds.ReadOnly,
		UID:               ds.Uid,
	}

	return &dsCfg, nil
}
