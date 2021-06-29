package configreader

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/services/vcs/vcsmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO what do we do with typeName and readOnly (expected editable)
// TODO test with multiple values in latest map? The order won't be guaranteed though.

func Test_vcsConfigReader_ReadConfigs(t *testing.T) {
	type args struct {
	}
	tt := []struct {
		name    string
		latest  map[string]vcs.VersionedObject
		configs []*datasources.Configs
		wantErr error
	}{
		{
			name:    "should work with empty latest",
			latest:  nil,
			configs: []*datasources.Configs{},
			wantErr: nil,
		},
		{
			name: "should convert a Postgres Datasrc versioned object",
			latest: map[string]vcs.VersionedObject{
				"postgresDatasrc": {
					ID:      "postgresDatasrc",
					Version: "test_version",
					Kind:    vcs.Datasource,
					Data: []byte(`{
						    "uid": "postgresDatasrc",
						    "orgId": 1,
						    "name": "PostgreSQL",
						    "type": "postgres",
						    "typeName": "PostgreSQL",
						    "access": "proxy",
						    "url": "localhost:5432",
						    "password": "user",
						    "user": "user",
						    "database": "database",
						    "basicAuth": false,
						    "isDefault": false,
						    "jsonData": {
						      "postgresVersion": 903,
						      "sslmode": "disable",
						      "tlsAuth": false,
						      "tlsAuthWithCACert": false,
						      "tlsConfigurationMethod": "file-path",
						      "tlsSkipVerify": true
						    },
						    "readOnly": false
						  }`),
					Timestamp: 0,
				},
			},
			configs: []*datasources.Configs{
				{
					APIVersion: 1,
					Datasources: []*datasources.UpsertDataSourceFromConfig{
						{
							OrgID:     1,
							Name:      "PostgreSQL",
							Type:      "postgres",
							Access:    "proxy",
							URL:       "localhost:5432",
							Password:  "user",
							User:      "user",
							Database:  "database",
							BasicAuth: false,
							IsDefault: false,
							JSONData: map[string]interface{}{
								"postgresVersion":        float64(903),
								"sslmode":                "disable",
								"tlsAuth":                false,
								"tlsAuthWithCACert":      false,
								"tlsConfigurationMethod": "file-path",
								"tlsSkipVerify":          true,
							},
							UID: "postgresDatasrc",
						},
					},
					DeleteDatasources: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "should convert a Prometheus Datasrc versioned object",
			latest: map[string]vcs.VersionedObject{
				"prometheusDatasrc": {
					ID:      "prometheusDatasrc",
					Version: "test_version",
					Kind:    vcs.Datasource,
					Data: []byte(`{
						    "uid": "prometheusDatasrc",
						    "orgId": 2,
						    "name": "Prometheus",
						    "type": "prometheus",
						    "access": "proxy",
						    "url": "localhost:9090"
						  }`),
					Timestamp: 0,
				},
			},
			configs: []*datasources.Configs{
				{
					APIVersion: 1,
					Datasources: []*datasources.UpsertDataSourceFromConfig{
						{
							OrgID:  2,
							Name:   "Prometheus",
							Type:   "prometheus",
							Access: "proxy",
							URL:    "localhost:9090",
							UID:    "prometheusDatasrc",
						},
					},
					DeleteDatasources: nil,
				},
			},
			wantErr: nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			calls := vcsmock.Calls{}
			vcsMock := vcsmock.VCSServiceMock{Calls: &calls}
			vcsMock.LatestFunc = func(c context.Context, k vcs.Kind) (map[string]vcs.VersionedObject, error) {
				return tc.latest, nil
			}

			cr := &vcsConfigReader{
				log: log.New("accesscontrol.provisioner-test"),
				vcs: &vcsMock,
			}

			readCfgs, err := cr.ReadConfigs(context.TODO())
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			assert.EqualValues(t, tc.configs, readCfgs)
		})
	}
}
