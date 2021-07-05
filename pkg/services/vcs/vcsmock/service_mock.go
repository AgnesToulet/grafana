package vcsmock

import (
	"context"

	"github.com/grafana/grafana/pkg/services/vcs"
)

var _ vcs.Service = &VCSServiceMock{}

type Calls struct {
	Store   []interface{}
	Latest  []interface{}
	History []interface{}
	Get     []interface{}
}

// VCSServiceMock is a mock implementation of the VCS Service
type VCSServiceMock struct {
	Calls       *Calls
	StoreFunc   func(context.Context, vcs.VersionedObject) (*vcs.VersionedObject, error)
	LatestFunc  func(context.Context, vcs.Kind) (map[string]vcs.VersionedObject, error)
	HistoryFunc func(context.Context, vcs.Kind, string) ([]vcs.VersionedObject, error)
	GetFunc     func(context.Context, vcs.Kind, string, string) (*vcs.VersionedObject, error)
}

func (m *VCSServiceMock) Store(ctx context.Context, vObj vcs.VersionedObject) (*vcs.VersionedObject, error) {
	m.Calls.Store = append(m.Calls.Store, []interface{}{ctx, vObj})
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, vObj)
	}
	return nil, nil
}

func (m *VCSServiceMock) Latest(ctx context.Context, kind vcs.Kind) (map[string]vcs.VersionedObject, error) {
	m.Calls.Latest = append(m.Calls.Latest, []interface{}{ctx, kind})
	if m.LatestFunc != nil {
		return m.LatestFunc(ctx, kind)
	}
	return nil, nil
}

func (m *VCSServiceMock) History(ctx context.Context, kind vcs.Kind, ID string) ([]vcs.VersionedObject, error) {
	m.Calls.History = append(m.Calls.History, []interface{}{ctx, kind, ID})
	if m.HistoryFunc != nil {
		return m.HistoryFunc(ctx, kind, ID)
	}
	return nil, nil
}

func (m *VCSServiceMock) Get(ctx context.Context, kind vcs.Kind, ID string, version string) (*vcs.VersionedObject, error) {
	m.Calls.Get = append(m.Calls.Get, []interface{}{ctx, kind, ID, version})
	if m.GetFunc != nil {
		return m.GetFunc(ctx, kind, ID, version)
	}
	return nil, nil
}
