package configreader

import (
	"context"

	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
)

type ConfigReader interface {
	ReadConfigs(context.Context) ([]*datasources.Configs, error)
}
