package configreader

import (
	"context"
	"os"
	"testing"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	logger log.Logger = log.New("fake.log")

	allProperties = "disktestdata/all-properties"
	versionZero   = "disktestdata/version-0"
	brokenYaml    = "disktestdata/broken-yaml"
	invalidAccess = "disktestdata/invalid-access"
)

func TestDatasourceAsConfig(t *testing.T) {
	Convey("Testing datasource as configuration", t, func() {
		bus.ClearBusHandlers()
		bus.AddHandler("test", mockGetOrg)

		Convey("broken yaml should return error", func() {
			reader := &diskConfigReader{configPath: brokenYaml}
			_, err := reader.ReadConfigs(context.TODO())
			So(err, ShouldNotBeNil)
		})

		Convey("invalid access should warn about invalid value and return 'proxy'", func() {
			reader := &diskConfigReader{log: logger, configPath: invalidAccess}
			configs, err := reader.ReadConfigs(context.TODO())
			So(err, ShouldBeNil)
			So(configs[0].Datasources[0].Access, ShouldEqual, models.DS_ACCESS_PROXY)
		})

		Convey("skip invalid directory", func() {
			cfgProvider := &diskConfigReader{log: log.New("test logger"), configPath: "./invalid-directory"}
			cfg, err := cfgProvider.ReadConfigs(context.TODO())
			if err != nil {
				t.Fatalf("ReadConfig returns an error %v", err)
			}

			So(len(cfg), ShouldEqual, 0)
		})

		Convey("can read all properties from version 1", func() {
			_ = os.Setenv("TEST_VAR", "name")
			cfgProvider := &diskConfigReader{log: log.New("test logger"), configPath: allProperties}
			cfg, err := cfgProvider.ReadConfigs(context.TODO())
			_ = os.Unsetenv("TEST_VAR")
			if err != nil {
				t.Fatalf("ReadConfig returns an error %v", err)
			}

			So(len(cfg), ShouldEqual, 3)

			dsCfg := cfg[0]

			So(dsCfg.APIVersion, ShouldEqual, 1)

			validateDatasourceV1(dsCfg)
			validateDeleteDatasources(dsCfg)

			dsCount := 0
			delDsCount := 0

			for _, c := range cfg {
				dsCount += len(c.Datasources)
				delDsCount += len(c.DeleteDatasources)
			}

			So(dsCount, ShouldEqual, 2)
			So(delDsCount, ShouldEqual, 1)
		})

		Convey("can read all properties from version 0", func() {
			cfgProvider := &diskConfigReader{log: log.New("test logger"), configPath: versionZero}
			cfg, err := cfgProvider.ReadConfigs(context.TODO())
			if err != nil {
				t.Fatalf("ReadConfig returns an error %v", err)
			}

			So(len(cfg), ShouldEqual, 1)

			dsCfg := cfg[0]

			So(dsCfg.APIVersion, ShouldEqual, 0)

			validateDatasource(dsCfg)
			validateDeleteDatasources(dsCfg)
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

func mockGetOrg(_ *models.GetOrgByIdQuery) error {
	return nil
}
