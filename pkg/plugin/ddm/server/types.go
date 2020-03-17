package server

import (
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
)

const (
	EngineName      = "engineName"
	EngineNamespace = "engineNamespace"
	SnapshotName    = "snapshotName"
)

type DDM interface {
	Init(plugin string, engine v1alpha1.MoveEngine) error
	SyncData(plugin string, ds v1alpha1.DataSync) (string, error)
	SyncStatus(plugin string, ds v1alpha1.DataSync) (int32, error)
}
