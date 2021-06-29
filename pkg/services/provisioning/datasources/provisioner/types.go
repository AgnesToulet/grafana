package provisioner

import (
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/provisioning/datasources"
)

func createInsertCommand(ds *datasources.UpsertDataSourceFromConfig) *models.AddDataSourceCommand {
	jsonData := simplejson.New()
	if len(ds.JSONData) > 0 {
		for k, v := range ds.JSONData {
			jsonData.Set(k, v)
		}
	}

	return &models.AddDataSourceCommand{
		OrgId:             ds.OrgID,
		Name:              ds.Name,
		Type:              ds.Type,
		Access:            models.DsAccess(ds.Access),
		Url:               ds.URL,
		Password:          ds.Password,
		User:              ds.User,
		Database:          ds.Database,
		BasicAuth:         ds.BasicAuth,
		BasicAuthUser:     ds.BasicAuthUser,
		BasicAuthPassword: ds.BasicAuthPassword,
		WithCredentials:   ds.WithCredentials,
		IsDefault:         ds.IsDefault,
		JsonData:          jsonData,
		SecureJsonData:    ds.SecureJSONData,
		ReadOnly:          !ds.Editable,
		Uid:               ds.UID,
	}
}

func createUpdateCommand(ds *datasources.UpsertDataSourceFromConfig, id int64) *models.UpdateDataSourceCommand {
	jsonData := simplejson.New()
	if len(ds.JSONData) > 0 {
		for k, v := range ds.JSONData {
			jsonData.Set(k, v)
		}
	}

	return &models.UpdateDataSourceCommand{
		Id:                id,
		Uid:               ds.UID,
		OrgId:             ds.OrgID,
		Name:              ds.Name,
		Type:              ds.Type,
		Access:            models.DsAccess(ds.Access),
		Url:               ds.URL,
		Password:          ds.Password,
		User:              ds.User,
		Database:          ds.Database,
		BasicAuth:         ds.BasicAuth,
		BasicAuthUser:     ds.BasicAuthUser,
		BasicAuthPassword: ds.BasicAuthPassword,
		WithCredentials:   ds.WithCredentials,
		IsDefault:         ds.IsDefault,
		JsonData:          jsonData,
		SecureJsonData:    ds.SecureJSONData,
		ReadOnly:          !ds.Editable,
	}
}
