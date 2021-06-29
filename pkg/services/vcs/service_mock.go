package vcs

import (
	"context"
)

var _ Service = &VCSServiceMock{}

type Calls struct {
	Store   []interface{}
	Latest  []interface{}
	History []interface{}
}

// VCSServiceMock is a mock implementation of the VCS Service
type VCSServiceMock struct {
	Calls       *Calls
	StoreFunc   func(context.Context, VersionedObject) error
	LatestFunc  func(context.Context, Kind) (map[string]VersionedObject, error)
	HistoryFunc func(context.Context, Kind, string) ([]VersionedObject, error)
}

func (m *VCSServiceMock) Store(ctx context.Context, vObj VersionedObject) error {
	m.Calls.Store = append(m.Calls.Store, []interface{}{ctx, vObj})
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, vObj)
	}
	return nil
}

func (m *VCSServiceMock) Latest(ctx context.Context, kind Kind) (map[string]VersionedObject, error) {
	m.Calls.Latest = append(m.Calls.Latest, []interface{}{ctx, kind})
	if m.LatestFunc != nil {
		return m.LatestFunc(ctx, kind)
	}
	return nil, nil
}

func (m *VCSServiceMock) History(ctx context.Context, kind Kind, ID string) ([]VersionedObject, error) {
	m.Calls.History = append(m.Calls.History, []interface{}{ctx, kind, ID})
	if m.HistoryFunc != nil {
		return m.HistoryFunc(ctx, kind, ID)
	}
	return nil, nil
}
