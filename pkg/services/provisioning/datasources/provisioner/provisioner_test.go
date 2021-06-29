package provisioner

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources/configreader"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/services/vcs/vcsmock"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	logger log.Logger = log.New("fake.log")

	twoDatasourcesConfig            = "../configreader/disktestdata/two-datasources"
	twoDatasourcesConfigPurgeOthers = "../configreader/disktestdata/insert-two-delete-two"
	doubleDatasourcesConfig         = "../configreader/disktestdata/double-default"
	multipleOrgsWithDefault         = "../configreader/disktestdata/multiple-org-default"
	withoutDefaults                 = "../configreader/disktestdata/appliedDefaults"

	fakeRepo *fakeRepository
)

func setupTestEnv(t testing.TB, gitops bool, configPath string) (*DatasourceProvisioner, *vcsmock.VCSServiceMock) {
	t.Helper()
	logger = log.New("datasource.provisioner-test")

	// Set gitops feature toggle
	cfg := setting.NewCfg()
	cfg.FeatureToggles = map[string]bool{"gitops": gitops}

	calls := vcsmock.Calls{}
	vcsMock := vcsmock.VCSServiceMock{Calls: &calls}

	provisioner := NewDatasourceProvisioner(cfg, &vcsMock)
	// Override provisioner's logger
	provisioner.log = logger

	// Override the config reader to use test data and test logger
	if !gitops {
		provisioner.cfgProvider = configreader.NewDiskConfigReader(logger, configPath)
	}

	return &provisioner, &vcsMock
}

func setupBusMock(t testing.TB) {
	t.Helper()

	fakeRepo = &fakeRepository{}
	bus.ClearBusHandlers()
	bus.AddHandler("test", mockDelete)
	bus.AddHandler("test", mockInsert)
	bus.AddHandler("test", mockUpdate)
	bus.AddHandler("test", mockGet)
	bus.AddHandler("test", mockGetOrg)
}

func TestDatasourceAsConfigFromFile(t *testing.T) {

	Convey("Testing datasource as configuration", t, func() {
		setupBusMock(t)

		Convey("apply default values when missing", func() {
			dc, _ := setupTestEnv(t, false, withoutDefaults)
			err := dc.applyChanges(context.TODO())
			if err != nil {
				t.Fatalf("applyChanges return an error %v", err)
			}

			So(len(fakeRepo.inserted), ShouldEqual, 1)
			So(fakeRepo.inserted[0].OrgId, ShouldEqual, 1)
			So(fakeRepo.inserted[0].Access, ShouldEqual, "proxy")
		})

		Convey("One configured datasource", func() {
			Convey("no datasource in database", func() {
				dc, _ := setupTestEnv(t, false, twoDatasourcesConfig)
				err := dc.applyChanges(context.TODO())
				if err != nil {
					t.Fatalf("applyChanges return an error %v", err)
				}

				So(len(fakeRepo.deleted), ShouldEqual, 0)
				So(len(fakeRepo.inserted), ShouldEqual, 2)
				So(len(fakeRepo.updated), ShouldEqual, 0)
			})

			Convey("One datasource in database with same name", func() {
				fakeRepo.loadAll = []*models.DataSource{
					{Name: "Graphite", OrgId: 1, Id: 1},
				}

				Convey("should update one datasource", func() {
					dc, _ := setupTestEnv(t, false, twoDatasourcesConfig)
					err := dc.applyChanges(context.TODO())
					if err != nil {
						t.Fatalf("applyChanges return an error %v", err)
					}

					So(len(fakeRepo.deleted), ShouldEqual, 0)
					So(len(fakeRepo.inserted), ShouldEqual, 1)
					So(len(fakeRepo.updated), ShouldEqual, 1)
				})
			})

			Convey("Two datasources with is_default", func() {
				dc, _ := setupTestEnv(t, false, doubleDatasourcesConfig)
				err := dc.applyChanges(context.TODO())
				Convey("should raise error", func() {
					So(err, ShouldEqual, datasources.ErrInvalidConfigToManyDefault)
				})
			})
		})

		Convey("Multiple datasources in different organizations with isDefault in each organization", func() {
			dc, _ := setupTestEnv(t, false, multipleOrgsWithDefault)
			err := dc.applyChanges(context.TODO())
			Convey("should not raise error", func() {
				So(err, ShouldBeNil)
				So(len(fakeRepo.inserted), ShouldEqual, 4)
				So(fakeRepo.inserted[0].IsDefault, ShouldBeTrue)
				So(fakeRepo.inserted[0].OrgId, ShouldEqual, 1)
				So(fakeRepo.inserted[2].IsDefault, ShouldBeTrue)
				So(fakeRepo.inserted[2].OrgId, ShouldEqual, 2)
			})
		})

		Convey("Two configured datasource and purge others ", func() {
			Convey("two other datasources in database", func() {
				fakeRepo.loadAll = []*models.DataSource{
					{Name: "old-graphite", OrgId: 1, Id: 1},
					{Name: "old-graphite2", OrgId: 1, Id: 2},
				}

				Convey("should have two new datasources", func() {
					dc, _ := setupTestEnv(t, false, twoDatasourcesConfigPurgeOthers)
					err := dc.applyChanges(context.TODO())
					if err != nil {
						t.Fatalf("applyChanges return an error %v", err)
					}

					So(len(fakeRepo.deleted), ShouldEqual, 2)
					So(len(fakeRepo.inserted), ShouldEqual, 2)
					So(len(fakeRepo.updated), ShouldEqual, 0)
				})
			})
		})

		Convey("Two configured datasource and purge others = false", func() {
			Convey("two other datasources in database", func() {
				fakeRepo.loadAll = []*models.DataSource{
					{Name: "Graphite", OrgId: 1, Id: 1},
					{Name: "old-graphite2", OrgId: 1, Id: 2},
				}

				Convey("should have two new datasources", func() {
					dc, _ := setupTestEnv(t, false, twoDatasourcesConfig)
					err := dc.applyChanges(context.TODO())
					if err != nil {
						t.Fatalf("applyChanges return an error %v", err)
					}

					So(len(fakeRepo.deleted), ShouldEqual, 0)
					So(len(fakeRepo.inserted), ShouldEqual, 1)
					So(len(fakeRepo.updated), ShouldEqual, 1)
				})
			})
		})
	})
}

func validateDeleteDatasources(dsCfg *datasources.Configs) {
	So(len(dsCfg.DeleteDatasources), ShouldEqual, 1)
	deleteDs := dsCfg.DeleteDatasources[0]
	So(deleteDs.Name, ShouldEqual, "old-graphite3")
	So(deleteDs.OrgID, ShouldEqual, 2)
}

func validateDatasource(dsCfg *datasources.Configs) {
	ds := dsCfg.Datasources[0]
	So(ds.Name, ShouldEqual, "name")
	So(ds.Type, ShouldEqual, "type")
	So(ds.Access, ShouldEqual, models.DS_ACCESS_PROXY)
	So(ds.OrgID, ShouldEqual, 2)
	So(ds.URL, ShouldEqual, "url")
	So(ds.User, ShouldEqual, "user")
	So(ds.Password, ShouldEqual, "password")
	So(ds.Database, ShouldEqual, "database")
	So(ds.BasicAuth, ShouldBeTrue)
	So(ds.BasicAuthUser, ShouldEqual, "basic_auth_user")
	So(ds.BasicAuthPassword, ShouldEqual, "basic_auth_password")
	So(ds.WithCredentials, ShouldBeTrue)
	So(ds.IsDefault, ShouldBeTrue)
	So(ds.Editable, ShouldBeTrue)
	So(ds.Version, ShouldEqual, 10)

	So(len(ds.JSONData), ShouldBeGreaterThan, 2)
	So(ds.JSONData["graphiteVersion"], ShouldEqual, "1.1")
	So(ds.JSONData["tlsAuth"], ShouldEqual, true)
	So(ds.JSONData["tlsAuthWithCACert"], ShouldEqual, true)

	So(len(ds.SecureJSONData), ShouldBeGreaterThan, 2)
	So(ds.SecureJSONData["tlsCACert"], ShouldEqual, "MjNOcW9RdkbUDHZmpco2HCYzVq9dE+i6Yi+gmUJotq5CDA==")
	So(ds.SecureJSONData["tlsClientCert"], ShouldEqual, "ckN0dGlyMXN503YNfjTcf9CV+GGQneN+xmAclQ==")
	So(ds.SecureJSONData["tlsClientKey"], ShouldEqual, "ZkN4aG1aNkja/gKAB1wlnKFIsy2SRDq4slrM0A==")
}

func validateDatasourceV1(dsCfg *datasources.Configs) {
	validateDatasource(dsCfg)
	ds := dsCfg.Datasources[0]
	So(ds.UID, ShouldEqual, "test_uid")
}

type BusActionCount struct {
	insertCnt, updateCnt, deleteCnt int
}

func TestProvisionFromVCS(t *testing.T) {
	tt := []struct {
		name    string
		latest  map[string]vcs.VersionedObject
		loadAll []*models.DataSource
		busCnt  BusActionCount
		wantErr error
	}{
		{
			name: "should work with empty latest",
		},
		{
			name: "should insert a Postgres Datasrc",
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
			loadAll: []*models.DataSource{},
			busCnt:  BusActionCount{insertCnt: 1},
			wantErr: nil,
		},
		{
			name: "should update a Prometheus Datasrc",
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
			loadAll: []*models.DataSource{{Id: 1, OrgId: 2, Name: "Prometheus"}},
			busCnt:  BusActionCount{updateCnt: 1},
			wantErr: nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test env
			setupBusMock(t)
			dc, vcsMock := setupTestEnv(t, true, "")
			vcsMock.LatestFunc = func(c context.Context, k vcs.Kind) (map[string]vcs.VersionedObject, error) {
				return tc.latest, nil
			}
			fakeRepo.loadAll = tc.loadAll

			// Provision
			err := dc.Provision(context.TODO())

			// Check result
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			assert.Len(t, fakeRepo.deleted, tc.busCnt.deleteCnt)
			assert.Len(t, fakeRepo.inserted, tc.busCnt.insertCnt)
			assert.Len(t, fakeRepo.updated, tc.busCnt.updateCnt)
		})
	}
}

type fakeRepository struct {
	inserted []*models.AddDataSourceCommand
	deleted  []*models.DeleteDataSourceCommand
	updated  []*models.UpdateDataSourceCommand

	loadAll []*models.DataSource
}

func mockDelete(cmd *models.DeleteDataSourceCommand) error {
	fakeRepo.deleted = append(fakeRepo.deleted, cmd)
	return nil
}

func mockUpdate(cmd *models.UpdateDataSourceCommand) error {
	fakeRepo.updated = append(fakeRepo.updated, cmd)
	return nil
}

func mockInsert(cmd *models.AddDataSourceCommand) error {
	fakeRepo.inserted = append(fakeRepo.inserted, cmd)
	return nil
}

func mockGet(cmd *models.GetDataSourceQuery) error {
	for _, v := range fakeRepo.loadAll {
		if cmd.Name == v.Name && cmd.OrgId == v.OrgId {
			cmd.Result = v
			return nil
		}
	}

	return models.ErrDataSourceNotFound
}

func mockGetOrg(_ *models.GetOrgByIdQuery) error {
	return nil
}
