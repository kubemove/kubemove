package server

import (
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
)

type DDM interface {
	Init(plugin string, config map[string]string) error
	SyncData(plugin string, ds v1alpha1.DataSync) (string, error)
	SyncStatus(plugin, id string) (int32, error)
}
