package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
)

func (m *MoveEngineAction) CreateDataSync() (string, error) {
	dsObj := &v1alpha1.DataSync{}

	meSpec := m.MEngine.Spec

	dsObj.Namespace = m.MEngine.Namespace
	dsObj.Name = fmt.Sprintf("ds-%s-%s", m.MEngine.Name, time.Now().Format("20060102150405"))
	spec := v1alpha1.DataSyncSpec{
		Namespace:      meSpec.Namespace,
		PluginProvider: meSpec.PluginProvider,
		MoveEngine:     m.MEngine.Name,
		Backup:         true,
	}

	for _, l := range m.syncedVolMap {
		vObj := &v1alpha1.DataVolume{
			Namespace:       l.Namespace,
			Name:            l.Volume,
			PVC:             l.PVC,
			RemoteName:      l.RemoteVolume,
			RemoteNamespace: l.RemoteNamespace,
		}
		spec.Volume = append(spec.Volume, vObj)
	}

	dsObj.Spec = spec

	err := m.client.Create(context.TODO(), dsObj)
	if err != nil {
		return "", err
	}
	return dsObj.Name, nil
}
