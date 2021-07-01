package vcs

import (
	"context"
	"fmt"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/backendplugin/pluginextensionv2"
	"github.com/grafana/grafana/pkg/registry"
)

const ServiceName = "VCSService"

type VCSService struct {
	PluginManager plugins.Manager `inject:""`

	log    log.Logger
	plugin *plugins.VCSPlugin
}

func init() {
	// Start service after Plugin Manager
	registry.Register(&registry.Descriptor{
		Name:         ServiceName,
		Instance:     &VCSService{},
		InitPriority: registry.MediumHigh - 1,
	})
}

func (vs *VCSService) Init() error {
	vs.log = log.New("vcs plugin")

	vs.plugin = vs.PluginManager.VersionedControlStorage()

	if vs.plugin != nil {
		if err := vs.plugin.Start(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (vs *VCSService) Store(ctx context.Context, object VersionedObject) error {
	if vs.plugin == nil {
		vs.log.Warn("VCS plugin has not been instantiated correctly")
		return nil
	}

	req := &pluginextensionv2.StoreRequest{
		VersionedObject: toPluginVersionedObject(object),
	}

	resp, err := vs.plugin.GRPCPlugin.Store(ctx, req)
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return fmt.Errorf("storing into versioned control storage failed: %s", resp.Error)
	}

	return nil
}

func (vs *VCSService) Latest(ctx context.Context, kind Kind) (map[string]VersionedObject, error) {
	if vs.plugin == nil {
		vs.log.Warn("VCS plugin has not been instantiated correctly")
		return nil, nil
	}

	req := &pluginextensionv2.LatestRequest{
		Kind: string(kind),
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

func (vs *VCSService) History(ctx context.Context, kind Kind, ID string) ([]VersionedObject, error) {
	if vs.plugin == nil {
		vs.log.Warn("VCS plugin has not been instantiated correctly")
		return nil, nil
	}

	req := &pluginextensionv2.HistoryRequest{
		Kind: string(kind),
		Id:   ID,
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
