package configreader

import "github.com/grafana/grafana/pkg/services/provisioning/datasources"

type ConfigReader interface {
	ReadConfigs() ([]*datasources.Configs, error)
}
