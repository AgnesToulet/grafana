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
	registry.Register(&registry.Descriptor{
		Name:         ServiceName,
		Instance:     &VCSService{},
		InitPriority: registry.High,
	})
}

func (vs *VCSService) Init() error {
	vs.log = log.New("vcs plugin")

	return nil
}

func (vs *VCSService) Run(ctx context.Context) error {
	vs.plugin = vs.PluginManager.VersionedControlStorage()

	if err := vs.plugin.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (vs *VCSService) Store(ctx context.Context, object VersionedObject) error {
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
