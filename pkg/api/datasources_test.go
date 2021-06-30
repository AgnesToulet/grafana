package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/services/vcs"
	"github.com/grafana/grafana/pkg/services/vcs/vcsmock"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	testOrgID     int64  = 1
	testUserID    int64  = 1
	testUserLogin string = "testUser"
)

func TestDataSourcesProxy_userLoggedIn(t *testing.T) {
	loggedInUserScenario(t, "When calling GET on", "/api/datasources/", func(sc *scenarioContext) {
		// Stubs the database query
		bus.AddHandler("test", func(query *models.GetDataSourcesQuery) error {
			assert.Equal(t, testOrgID, query.OrgId)
			query.Result = []*models.DataSource{
				{Name: "mmm"},
				{Name: "ZZZ"},
				{Name: "BBB"},
				{Name: "aaa"},
			}
			return nil
		})

		// handler func being tested
		hs := &HTTPServer{
			Bus:           bus.GetBus(),
			Cfg:           setting.NewCfg(),
			PluginManager: &fakePluginManager{},
		}
		sc.handlerFunc = hs.GetDataSources
		sc.fakeReq("GET", "/api/datasources").exec()

		respJSON := []map[string]interface{}{}
		err := json.NewDecoder(sc.resp.Body).Decode(&respJSON)
		require.NoError(t, err)

		assert.Equal(t, "aaa", respJSON[0]["name"])
		assert.Equal(t, "BBB", respJSON[1]["name"])
		assert.Equal(t, "mmm", respJSON[2]["name"])
		assert.Equal(t, "ZZZ", respJSON[3]["name"])
	})

	loggedInUserScenario(t, "Should be able to save a data source when calling DELETE on non-existing",
		"/api/datasources/name/12345", func(sc *scenarioContext) {
			// handler func being tested
			hs := &HTTPServer{
				Bus:           bus.GetBus(),
				Cfg:           setting.NewCfg(),
				PluginManager: &fakePluginManager{},
			}
			sc.handlerFunc = hs.DeleteDataSourceByName
			sc.fakeReqWithParams("DELETE", sc.url, map[string]string{}).exec()
			assert.Equal(t, 404, sc.resp.Code)
		})
}

// Adding data sources with invalid URLs should lead to an error.
func TestAddDataSource_InvalidURL(t *testing.T) {
	defer bus.ClearBusHandlers()
	// handler func being tested
	hs := &HTTPServer{
		Bus: bus.GetBus(),
		Cfg: setting.NewCfg(),
	}

	sc := setupScenarioContext(t, "/api/datasources")

	sc.m.Post(sc.url, routing.Wrap(func(c *models.ReqContext) response.Response {
		return hs.AddDataSource(c, models.AddDataSourceCommand{
			Name: "Test",
			Url:  "invalid:url",
		})
	}))

	sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()

	assert.Equal(t, 400, sc.resp.Code)
}

func TestVCSStoreDataSource(t *testing.T) {
	defer bus.ClearBusHandlers()

	db := sqlstore.InitTestDB(t)
	cfg := setting.NewCfg()
	cfg.FeatureToggles = map[string]bool{"gitops": true}
	calls := vcsmock.Calls{}
	vcsMock := vcsmock.VCSServiceMock{Calls: &calls}
	hs := &HTTPServer{
		Bus:      bus.GetBus(),
		Cfg:      cfg,
		SQLStore: db,
		VCS:      &vcsMock,
	}

	sc := setupScenarioContext(t, "/api/datasources")

	addCmd := models.AddDataSourceCommand{
		Name:           "PostgresDatasrc",
		Type:           "postresql",
		Access:         "proxy",
		Url:            "localhost:5432",
		Database:       "testDatabase",
		User:           testUserLogin,
		BasicAuth:      false,
		IsDefault:      false,
		OrgId:          testOrgID,
		ReadOnly:       false,
		Result:         &models.DataSource{},
		SecureJsonData: nil,
		JsonData: simplejson.NewFromAny(map[string]interface{}{
			"postgresVersion":        "903",
			"sslmode":                "disable",
			"tlsAuth":                false,
			"tlsAuthWithCACert":      false,
			"tlsConfigurationMethod": "file-path",
			"tlsSkipVerify":          true,
		}),
	}

	sc.m.Post(sc.url, routing.Wrap(
		func(c *models.ReqContext) response.Response {
			sc.context = c
			sc.context.SignedInUser = &models.SignedInUser{
				UserId: testUserID,
				OrgId:  testOrgID,
				Login:  testUserLogin,
			}
			return hs.AddDataSource(c, addCmd)
		}),
	)

	sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
	assert.Equal(t, 200, sc.resp.Code)

	require.Len(t, vcsMock.Calls.Store, 1)
	storeCall, ok := vcsMock.Calls.Store[0].([]interface{})
	require.True(t, ok, "expected multiple parameters in vcs Store call")
	require.Len(t, storeCall, 2, "expected 2 parameters in vcs Store call")
	vObj, ok := storeCall[1].(vcs.VersionedObject)
	require.True(t, ok, "expected second parameter of vcs Store call to be a VersionedObject")
	assert.NotEmpty(t, vObj.ID, "expected versioned object ID to be set to datasource UID")

	// TODO make the conversion better
	type ProvisionDatasourcesDTO struct {
		models.DataSource
		Editable bool `json:"editable"`
	}
	var ds ProvisionDatasourcesDTO

	err := json.Unmarshal(vObj.Data, &ds)
	require.NoError(t, err)
	assert.Equal(t, addCmd.Name, ds.Name, "expected vobj name to be correctly set")
	assert.Equal(t, addCmd.Type, ds.Type, "expected vobj type to be correctly set")
	assert.Equal(t, addCmd.Access, ds.Access, "expected vobj access to be correctly set")
	assert.Equal(t, addCmd.Url, ds.Url, "expected vobj url to be correctly set")
	assert.Equal(t, addCmd.Database, ds.Database, "expected vobj database to be correctly set")
	assert.Equal(t, addCmd.User, ds.User, "expected vobj user to be correctly set")
	assert.Equal(t, addCmd.BasicAuth, ds.BasicAuth, "expected vobj basic auth to be correctly set")
	assert.Equal(t, addCmd.IsDefault, ds.IsDefault, "expected vobj is default to be correctly set")
	assert.Equal(t, addCmd.OrgId, ds.OrgId, "expected vobj orgID to be correctly set")
	assert.Equal(t, addCmd.ReadOnly, ds.ReadOnly, "expected vobj readonly to be correctly set")
	assert.Len(t, ds.SecureJsonData, 0, "expected obj securejsondata to be empty")
	assert.EqualValues(t, addCmd.JsonData, ds.JsonData, "expected vobj json data to be correctly set")
}

// Adding data sources with URLs not specifying protocol should work.
func TestAddDataSource_URLWithoutProtocol(t *testing.T) {
	defer bus.ClearBusHandlers()
	hs := &HTTPServer{
		Bus: bus.GetBus(),
		Cfg: setting.NewCfg(),
	}

	const name = "Test"
	const url = "localhost:5432"

	// Stub handler
	bus.AddHandler("sql", func(cmd *models.AddDataSourceCommand) error {
		assert.Equal(t, name, cmd.Name)
		assert.Equal(t, url, cmd.Url)

		cmd.Result = &models.DataSource{}
		return nil
	})

	sc := setupScenarioContext(t, "/api/datasources")

	sc.m.Post(sc.url, routing.Wrap(func(c *models.ReqContext) response.Response {
		return hs.AddDataSource(c, models.AddDataSourceCommand{
			Name: name,
			Url:  url,
		})
	}))

	sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()

	assert.Equal(t, 200, sc.resp.Code)
}

// Updating data sources with invalid URLs should lead to an error.
func TestUpdateDataSource_InvalidURL(t *testing.T) {
	defer bus.ClearBusHandlers()
	hs := &HTTPServer{
		Bus: bus.GetBus(),
		Cfg: setting.NewCfg(),
	}

	sc := setupScenarioContext(t, "/api/datasources/1234")

	sc.m.Put(sc.url, routing.Wrap(func(c *models.ReqContext) response.Response {
		return hs.AddDataSource(c, models.AddDataSourceCommand{
			Name: "Test",
			Url:  "invalid:url",
		})
	}))

	sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()

	assert.Equal(t, 400, sc.resp.Code)
}

// Updating data sources with URLs not specifying protocol should work.
func TestUpdateDataSource_URLWithoutProtocol(t *testing.T) {
	defer bus.ClearBusHandlers()
	hs := &HTTPServer{
		Bus: bus.GetBus(),
		Cfg: setting.NewCfg(),
	}

	const name = "Test"
	const url = "localhost:5432"

	// Stub handler
	bus.AddHandler("sql", func(cmd *models.AddDataSourceCommand) error {
		assert.Equal(t, name, cmd.Name)
		assert.Equal(t, url, cmd.Url)

		cmd.Result = &models.DataSource{}
		return nil
	})

	sc := setupScenarioContext(t, "/api/datasources/1234")

	sc.m.Put(sc.url, routing.Wrap(func(c *models.ReqContext) response.Response {
		return hs.AddDataSource(c, models.AddDataSourceCommand{
			Name: name,
			Url:  url,
		})
	}))

	sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()

	assert.Equal(t, 200, sc.resp.Code)
}
