package vcs

import "context"

type Kind string

const (
	Datasource Kind = "datasource"
)

type VersionedObject struct {
	ID        string
	Version   string
	Kind      Kind
	Data      []byte
	Timestamp int64
}

type Service interface {
	Store(context.Context, VersionedObject) error
	Latest(context.Context, Kind) map[string]VersionedObject
	History(context.Context, Kind, string) []VersionedObject
}
