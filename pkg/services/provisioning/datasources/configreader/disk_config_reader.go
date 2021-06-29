package configreader

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
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

	err = validateDefaultUniqueness(cr.log, datasources)
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
