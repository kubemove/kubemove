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

	dsObj.Spec = spec

	err := m.client.Create(context.TODO(), dsObj)
	if err != nil {
		return "", err
	}
	return dsObj.Name, nil
}
