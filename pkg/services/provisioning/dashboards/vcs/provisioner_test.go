package vcs

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/services/vcs/vcsmock"
	"github.com/grafana/grafana/pkg/setting"
)

func setupTestEnv(t *testing.T) (Provisioner, *vcsmock.VCSServiceMock) {
	t.Helper()
	store := sqlstore.InitTestDB(t)

	config := setting.NewCfg()
	config.FeatureToggles = map[string]bool{"gitops": true}

	vcsMock := &vcsmock.VCSServiceMock{Calls: &vcsmock.Calls{}}
	vcsMock.IsDisabledFunc = func() bool { return false }

	p := Provisioner{
		VCS:   vcsMock,
		Store: store,
		Cfg:   config,
		log:   log.New("dashboards.provisioner-test"),
	}
	return p, vcsMock
}

func populateLatestFromFiles(t *testing.T, jsonFiles map[string]string, latest map[string]vcs.VersionedObject) {
	for dashUid, path := range jsonFiles {
		jsonFile, err := os.Open(path)
		defer jsonFile.Close()
		require.NoError(t, err)

		byteValue, err := ioutil.ReadAll(jsonFile)
		require.NoError(t, err)

		vobj := latest[dashUid]
		vobj.Data = byteValue
		latest[dashUid] = vobj
	}
}

func removeElementsFromLatest(t *testing.T, latest map[string]vcs.VersionedObject, toRemove map[string]empty) map[string]vcs.VersionedObject {
	newLatest := map[string]vcs.VersionedObject{}

	for k, v := range latest {
		if _, ok := toRemove[k]; !ok {
			newLatest[k] = v
		}
	}

	return newLatest
}

type empty struct{}

func TestProvisionAndCleanUpOrphanedFromVCS(t *testing.T) {
	tt := []struct {
		name      string
		jsonFiles map[string]string
		latest    map[string]vcs.VersionedObject
		toRemove  map[string]empty
		wantErr   error
	}{
		{
			name: "should work with empty latest",
		},
		{
			name:      "should insert a Panel",
			jsonFiles: map[string]string{"randomdash": "./testdata/random.json"},
			latest: map[string]vcs.VersionedObject{
				"randomdash": {
					ID:        "randomdash",
					Version:   "testsha",
					Kind:      vcs.Dashboard,
					Data:      []byte{},
					Timestamp: 0,
				},
			},
			toRemove: nil,
			wantErr:  nil,
		},
		{
			name: "should remove 1 Panel",
			jsonFiles: map[string]string{
				"randomdash":  "./testdata/random.json",
				"randomdash2": "./testdata/random2.json",
			},
			latest: map[string]vcs.VersionedObject{
				"randomdash": {
					ID:        "randomdash",
					Version:   "testsha",
					Kind:      vcs.Dashboard,
					Data:      []byte{},
					Timestamp: 0,
				},
				"randomdash2": {
					ID:        "randomdash2",
					Version:   "testsha2",
					Kind:      vcs.Dashboard,
					Data:      []byte{},
					Timestamp: 0,
				},
			},
			toRemove: map[string]empty{"randomdash": {}},
			wantErr:  nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test env
			dashProv, vcsMock := setupTestEnv(t)

			populateLatestFromFiles(t, tc.jsonFiles, tc.latest)

			vcsMock.LatestFunc = func(c context.Context, k vcs.Kind) (map[string]vcs.VersionedObject, error) {
				return tc.latest, nil
			}

			// Provision
			err := dashProv.Provision(context.TODO())

			// Check result
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			dashSvc := dashboards.NewProvisioningService(dashProv.Store)
			dashProvisioningInfo, err := dashSvc.GetProvisionedDashboardData(ProvisionerName)
			require.NoError(t, err)

			assert.Len(t, dashProvisioningInfo, len(tc.latest))

			// Test removal
			newLatest := removeElementsFromLatest(t, tc.latest, tc.toRemove)

			vcsMock.LatestFunc = func(c context.Context, k vcs.Kind) (map[string]vcs.VersionedObject, error) {
				return newLatest, nil
			}

			// Remove orphaned dashboards
			dashProv.CleanUpOrphanedDashboards()

			dashProvisioningInfo, err = dashSvc.GetProvisionedDashboardData(ProvisionerName)
			require.NoError(t, err)

			assert.Len(t, dashProvisioningInfo, len(tc.latest)-len(tc.toRemove))
		})
	}
}
