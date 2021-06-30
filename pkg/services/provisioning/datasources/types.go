package datasources

import (
	"context"
)

//ConfigReader reads datasource config files from storage
type ConfigReader interface {
	ReadConfigs(context.Context) ([]*Configs, error)
}

type Configs struct {
	APIVersion int64

	Datasources       []*UpsertDataSourceFromConfig
	DeleteDatasources []*DeleteDatasourceConfig
}

type DeleteDatasourceConfig struct {
	OrgID int64
	Name  string
}

type UpsertDataSourceFromConfig struct {
	OrgID   int64
	Version int

	Name              string
	Type              string
	Access            string
	URL               string
	Password          string
	User              string
	Database          string
	BasicAuth         bool
	BasicAuthUser     string
	BasicAuthPassword string
	WithCredentials   bool
	IsDefault         bool
	JSONData          map[string]interface{}
	SecureJSONData    map[string]string
	Editable          bool
	UID               string
}
