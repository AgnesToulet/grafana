package vcs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/infra/localcache"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/backendplugin/pluginextensionv2"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

const ServiceName = "PluginService"

type PluginService struct {
	PluginManager plugins.Manager          `inject:""`
	CacheService  *localcache.CacheService `inject:""`

	log    log.Logger
	plugin *plugins.VCSPlugin
}

func init() {
	// Start service after Plugin Manager
	registry.Register(&registry.Descriptor{
		Name:         ServiceName,
		Instance:     &PluginService{},
		InitPriority: registry.MediumHigh - 1,
	})
}

func (vs *PluginService) Init() error {
	vs.log = log.New("vcs plugin")

	vs.plugin = vs.PluginManager.VersionedControlStorage()

	if vs.plugin != nil {
		if err := vs.plugin.Start(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (vs *PluginService) Store(ctx context.Context, object VersionedObject) (*VersionedObject, error) {
	if vs.plugin == nil {
		vs.log.Warn("VCS plugin has not been instantiated correctly")
		return nil, nil
	}

	appInstanceSettings, err := vs.appInstanceSettings(vs.plugin.Id)
	if err != nil {
		return nil, err
	}

	req := &pluginextensionv2.StoreRequest{
		AppInstanceSettings: appInstanceSettings,
		VersionedObject:     toPluginVersionedObject(object),
	}

	resp, err := vs.plugin.GRPCPlugin.Store(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("storing into versioned control storage failed: %s", resp.Error)
	}

	vobj := fromPluginVersionedObject(resp.VersionedObject)
	return &vobj, nil
}

func (vs *PluginService) Latest(ctx context.Context, kind Kind) (map[string]VersionedObject, error) {
	if vs.plugin == nil {
		vs.log.Warn("VCS plugin has not been instantiated correctly")
		return nil, nil
	}

	appInstanceSettings, err := vs.appInstanceSettings(vs.plugin.Id)
	if err != nil {
		return nil, err
	}

	req := &pluginextensionv2.LatestRequest{
		AppInstanceSettings: appInstanceSettings,
		Kind:                string(kind),
	}

	resp, err := vs.plugin.GRPCPlugin.Latest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("retrieving latest version from versioned control storage failed: %s", resp.Error)
	}

	versionedObjects := make(map[string]VersionedObject, len(resp.LatestVersionedObjects))
	for key, val := range resp.LatestVersionedObjects {
		versionedObjects[key] = fromPluginVersionedObject(val)
	}
	return versionedObjects, nil
}

func (vs *PluginService) History(ctx context.Context, kind Kind, ID string) ([]VersionedObject, error) {
	if vs.plugin == nil {
		vs.log.Warn("VCS plugin has not been instantiated correctly")
		return nil, nil
	}
	appInstanceSettings, err := vs.appInstanceSettings(vs.plugin.Id)
	if err != nil {
		return nil, err
	}

	req := &pluginextensionv2.HistoryRequest{
		AppInstanceSettings: appInstanceSettings,
		Kind:                string(kind),
		Id:                  ID,
	}

	resp, err := vs.plugin.GRPCPlugin.History(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("retrieving latest version from versioned control storage failed: %s", resp.Error)
	}

	versionedObjects := make([]VersionedObject, len(resp.VersionedObjects))
	for i, obj := range resp.VersionedObjects {
		versionedObjects[i] = fromPluginVersionedObject(obj)
	}
	return versionedObjects, nil
}

const appSettingsMainOrg = 1
const appSettingsCacheTTL = 5 * time.Second
const appSettingsCachePrefix = "app-setting-"

func (vs *PluginService) appInstanceSettings(pluginID string) (*pluginextensionv2.AppInstanceSettings, error) {
	ps, err := vs.pluginSettings(pluginID)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(ps.JsonData)
	if err != nil {
		return nil, err
	}

	return &pluginextensionv2.AppInstanceSettings{
		JsonData:                jsonData,
		DecryptedSecureJsonData: ps.DecryptedValues(),
		LastUpdatedMS:           ps.Updated.Unix(),
	}, nil
}

func (vs *PluginService) pluginSettings(pluginID string) (*models.PluginSetting, error) {
	cacheKey := appSettingsCachePrefix + pluginID

	cached, found := vs.CacheService.Get(cacheKey)
	if found {
		return cached.(*models.PluginSetting), nil
	}

	q := &models.GetPluginSettingByIdQuery{
		PluginId: pluginID,
		OrgId:    appSettingsMainOrg,
	}

	if err := sqlstore.GetPluginSettingById(q); err != nil {
		return nil, err
	}

	vs.CacheService.Set(cacheKey, q.Result, appSettingsCacheTTL)

	return q.Result, nil
}

func toPluginVersionedObject(object VersionedObject) *pluginextensionv2.VersionedObject {
	return &pluginextensionv2.VersionedObject{
		Id:        object.ID,
		Version:   object.Version,
		Kind:      string(object.Kind),
		Data:      object.Data,
		Timestamp: object.Timestamp,
	}
}

func fromPluginVersionedObject(object *pluginextensionv2.VersionedObject) VersionedObject {
	return VersionedObject{
		ID:        object.Id,
		Version:   object.Version,
		Kind:      Kind(object.Kind),
		Data:      object.Data,
		Timestamp: object.Timestamp,
	}
}
