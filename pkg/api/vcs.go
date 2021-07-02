package api

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana/pkg/services/vcs"
)

func (hs *HTTPServer) storeObjInVCS(ctx context.Context, kind vcs.Kind, uid string, obj interface{}) error {
	if hs.VCS == nil || hs.VCS.IsDisabled() {
		return nil
	}

	dashJson, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	vobj := vcs.VersionedObject{
		ID:   uid,
		Kind: vcs.Dashboard,
		Data: dashJson,
	}

	_, err = hs.VCS.Store(ctx, vobj)

	return err
}
