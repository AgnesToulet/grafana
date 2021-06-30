package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/plugins/backendplugin"
	"github.com/grafana/grafana/pkg/plugins/backendplugin/grpcplugin"
	"github.com/grafana/grafana/pkg/plugins/backendplugin/pluginextensionv2"
	"github.com/grafana/grafana/pkg/util/errutil"
)

type VCSPlugin struct {
	FrontendPluginBase

	Executable           string `json:"executable,omitempty"`
	GRPCPlugin           pluginextensionv2.VCSPlugin
	backendPluginManager backendplugin.Manager
}

func (p *VCSPlugin) Load(decoder *json.Decoder, base *PluginBase, backendPluginManager backendplugin.Manager) (interface{}, error) {
	if err := decoder.Decode(p); err != nil {
		return nil, errutil.Wrapf(err, "Failed to decode versioned control storage plugin")
	}

	p.backendPluginManager = backendPluginManager

	cmd := ComposePluginStartCommand(p.Executable)
	fullpath := filepath.Join(base.PluginDir, cmd)
	factory := grpcplugin.NewVCSPlugin(p.Id, fullpath, p.onPluginStart)
	if err := backendPluginManager.RegisterAndStart(context.Background(), p.Id, factory); err != nil {
		return nil, errutil.Wrapf(err, "failed to register versioned control storage plugin")
	}

	return p, nil
}

func (p *VCSPlugin) Start(ctx context.Context) error {
	if err := p.backendPluginManager.StartPlugin(ctx, p.Id); err != nil {
		fmt.Println("err", err)
		return errutil.Wrapf(err, "Failed to start versioned control storage plugin")
	}

	return nil
}

func (p *VCSPlugin) onPluginStart(pluginID string, vcs pluginextensionv2.VCSPlugin, logger log.Logger) error {
	p.GRPCPlugin = vcs
	return nil
}
