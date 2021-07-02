package vcs

import "context"

type Kind string

const (
	Datasource Kind = "datasource"
)

type VersionedObject struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Kind      Kind   `json:"-"`
	Data      []byte `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

type Service interface {
	Store(context.Context, VersionedObject) (*VersionedObject, error)
	Latest(context.Context, Kind) (map[string]VersionedObject, error)
	History(context.Context, Kind, string) ([]VersionedObject, error)
}
