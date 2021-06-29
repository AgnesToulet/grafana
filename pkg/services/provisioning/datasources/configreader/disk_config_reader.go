package configreader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/provisioning/utils"
)

type diskConfigReader struct {
	log        log.Logger
	configPath string
}

func NewDiskConfigReader(log log.Logger, configPath string) datasources.ConfigReader {
	return &diskConfigReader{log: log, configPath: configPath}
}

func (cr *diskConfigReader) ReadConfigs(_ context.Context) ([]*datasources.Configs, error) {
	var datasources []*datasources.Configs

	files, err := ioutil.ReadDir(cr.configPath)
	if err != nil {
		cr.log.Error("can't read datasource provisioning files from directory", "path", cr.configPath, "error", err)
		return datasources, nil
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			datasource, err := cr.parseDatasourceConfig(cr.configPath, file)
			if err != nil {
				return nil, err
			}

			if datasource != nil {
				datasources = append(datasources, datasource)
			}
		}
	}

	err = cr.validateDefaultUniqueness(datasources)
	if err != nil {
		return nil, err
	}

	return datasources, nil
}

func (cr *diskConfigReader) parseDatasourceConfig(path string, file os.FileInfo) (*datasources.Configs, error) {
	filename, _ := filepath.Abs(filepath.Join(path, file.Name()))

	// nolint:gosec
	// We can ignore the gosec G304 warning on this one because `filename` comes from ps.Cfg.ProvisioningPath
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var apiVersion *configVersion
	err = yaml.Unmarshal(yamlFile, &apiVersion)
	if err != nil {
		return nil, err
	}

	if apiVersion == nil {
		apiVersion = &configVersion{APIVersion: 0}
	}

	if apiVersion.APIVersion > 0 {
		v1 := &configsV1{Log: cr.log}
		err = yaml.Unmarshal(yamlFile, v1)
		if err != nil {
			return nil, err
		}

		return v1.mapToDatasourceFromConfig(apiVersion.APIVersion), nil
	}

	var v0 *configsV0
	err = yaml.Unmarshal(yamlFile, &v0)
	if err != nil {
		return nil, err
	}

	cr.log.Warn("[Deprecated] the datasource provisioning config is outdated. please upgrade", "filename", filename)

	return v0.mapToDatasourceFromConfig(apiVersion.APIVersion), nil
}

func (cr *diskConfigReader) validateDefaultUniqueness(datasourcesCfg []*datasources.Configs) error {
	defaultCount := map[int64]int{}
	for i := range datasourcesCfg {
		if datasourcesCfg[i].Datasources == nil {
			continue
		}

		for _, ds := range datasourcesCfg[i].Datasources {
			if ds.OrgID == 0 {
				ds.OrgID = 1
			}

			if err := cr.validateAccessAndOrgID(ds); err != nil {
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

func (cr *diskConfigReader) validateAccessAndOrgID(ds *datasources.UpsertDataSourceFromConfig) error {
	if err := utils.CheckOrgExists(ds.OrgID); err != nil {
		return err
	}

	if ds.Access == "" {
		ds.Access = models.DS_ACCESS_PROXY
	}

	if ds.Access != models.DS_ACCESS_DIRECT && ds.Access != models.DS_ACCESS_PROXY {
		cr.log.Warn("invalid access value, will use 'proxy' instead", "value", ds.Access)
		ds.Access = models.DS_ACCESS_PROXY
	}
	return nil
}
