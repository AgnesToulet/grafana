package vcs

import (
	"context"

	"github.com/grafana/grafana/pkg/registry"
)

type Kind string

const (
	Datasource Kind = "datasource"
	Dashboard  Kind = "dashboard"
)

type VersionedObject struct {
	ID        string
	Version   string
	Kind      Kind
	Data      []byte
	Timestamp int64
}

type VersionedObjectDTO struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

type Service interface {
	registry.CanBeDisabled

	Store(context.Context, VersionedObject) (*VersionedObject, error)
	Latest(context.Context, Kind) (map[string]VersionedObject, error)
	History(context.Context, Kind, string) ([]VersionedObject, error)
	Get(context.Context, Kind, string, string) (*VersionedObject, error)
}
