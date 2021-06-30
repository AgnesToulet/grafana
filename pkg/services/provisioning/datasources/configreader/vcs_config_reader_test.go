package configreader

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/services/vcs/vcsmock"
)

func Test_vcsConfigReader_ReadConfigs(t *testing.T) {
	tt := []struct {
		name    string
		latest  map[string]vcs.VersionedObject
		configs *datasources.Configs
		wantErr error
	}{
		{
			name:    "should work with empty latest",
			latest:  nil,
			configs: nil,
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
						    "access": "proxy",
						    "url": "localhost:5432",
						    "password": "user",
						    "user": "user",
						    "database": "database",
						    "basicAuth": false,
						    "isDefault": false,
						    "editable": true,
						    "jsonData": {
						      "postgresVersion": 903,
						      "sslmode": "disable",
						      "tlsAuth": false,
						      "tlsAuthWithCACert": false,
						      "tlsConfigurationMethod": "file-path",
						      "tlsSkipVerify": true
						    }
						  }`),
					Timestamp: 0,
				},
			},
			configs: &datasources.Configs{
				APIVersion: 1,
				Datasources: []*datasources.UpsertDataSourceFromConfig{
					{
						OrgID:     1,
						UID:       "postgresDatasrc",
						Name:      "PostgreSQL",
						Type:      "postgres",
						Access:    "proxy",
						URL:       "localhost:5432",
						Password:  "user",
						User:      "user",
						Database:  "database",
						BasicAuth: false,
						IsDefault: false,
						Editable:  true,
						JSONData: map[string]interface{}{
							"postgresVersion":        json.Number("903"),
							"sslmode":                "disable",
							"tlsAuth":                false,
							"tlsAuthWithCACert":      false,
							"tlsConfigurationMethod": "file-path",
							"tlsSkipVerify":          true,
						},
					},
				},
				DeleteDatasources: nil,
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
				"prometheusDatasrc2": {
					ID:      "prometheusDatasrc2",
					Version: "test_version",
					Kind:    vcs.Datasource,
					Data: []byte(`{
						    "uid": "prometheusDatasrc2",
						    "orgId": 2,
						    "name": "Prometheus",
						    "type": "prometheus",
						    "access": "proxy",
						    "url": "localhost:9090"
						  }`),
					Timestamp: 0,
				},
			},
			configs: &datasources.Configs{
				APIVersion: 1,
				Datasources: []*datasources.UpsertDataSourceFromConfig{
					{
						OrgID:    2,
						Name:     "Prometheus",
						Type:     "prometheus",
						Access:   "proxy",
						URL:      "localhost:9090",
						UID:      "prometheusDatasrc",
						Editable: true, // TODO double check if this is what we want (default to true)
					},
					{
						OrgID:    2,
						Name:     "Prometheus",
						Type:     "prometheus",
						Access:   "proxy",
						URL:      "localhost:9090",
						UID:      "prometheusDatasrc2",
						Editable: true,
					},
				},
				DeleteDatasources: nil,
			},
			wantErr: nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Required for validation to pass
			bus.ClearBusHandlers()
			bus.AddHandler("test", mockGetOrg)

			// Setup
			calls := vcsmock.Calls{}
			vcsMock := vcsmock.VCSServiceMock{Calls: &calls}
			vcsMock.LatestFunc = func(c context.Context, k vcs.Kind) (map[string]vcs.VersionedObject, error) {
				return tc.latest, nil
			}
			cr := &vcsConfigReader{
				log: log.New("accesscontrol.provisioner-test"),
				vcs: &vcsMock,
			}

			// Test ReadConfigs
			readCfgs, err := cr.ReadConfigs(context.TODO())
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			if tc.configs == nil {
				require.Len(t, readCfgs, 0)
				return
			}

			require.Len(t, readCfgs, 1)
			assert.Equal(t, tc.configs.APIVersion, readCfgs[0].APIVersion)

			for _, ds := range tc.configs.Datasources {
				assert.Contains(t, readCfgs[0].Datasources, ds)
			}
			assert.Len(t, tc.configs.DeleteDatasources, 0)
		})
	}
}
