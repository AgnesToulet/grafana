package Service

import (
	"context"

	"github.com/grafana/grafana/pkg/services/vcs"
)

type Calls struct {
	Store   []interface{}
	Latest  []interface{}
	History []interface{}
}

// VCSServiceMock is a mock implementation of the VCS Service
type VCSServiceMock struct {
	Calls       *Calls
	StoreFunc   func(context.Context, vcs.VersionedObject) error
	LatestFunc  func(context.Context, vcs.Kind) (map[string]vcs.VersionedObject, error)
	HistoryFunc func(context.Context, vcs.Kind, string) ([]vcs.VersionedObject, error)
}

func (m *VCSServiceMock) Store(ctx context.Context, vObj vcs.VersionedObject) error {
	m.Calls.Store = append(m.Calls.Store, []interface{}{ctx, vObj})
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, vObj)
	}
	return nil
}

func (m *VCSServiceMock) Latest(ctx context.Context, kind vcs.Kind) (map[string]vcs.VersionedObject, error) {
	m.Calls.Latest = append(m.Calls.Latest, []interface{}{ctx, kind})
	if m.LatestFunc != nil {
		return m.LatestFunc(ctx, kind)
	}
	return nil, nil
}

func (m *VCSServiceMock) History(ctx context.Context, kind vcs.Kind, id string) ([]vcs.VersionedObject, error) {
	m.Calls.History = append(m.Calls.History, []interface{}{ctx, kind, id})
	if m.HistoryFunc != nil {
		return m.HistoryFunc(ctx, kind, id)
	}
	return nil, nil
}
