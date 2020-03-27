package engine

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

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
		Mode:           v1alpha1.SyncModeBackup,
	}

	dsObj.Spec = spec

	err := m.client.Create(context.TODO(), dsObj)
	if err != nil {
		return "", err
	}
	return dsObj.Name, nil
}

func (m *MoveEngineAction) CreateDataSyncAtRemote(dsName string) error {
	// find the source DataSync CR
	srcDS, err := GetDataSync(m.client, dsName, m.MEngine.Namespace)
	if err != nil {
		return err
	}

	newDs := &v1alpha1.DataSync{}

	newDs.Name = srcDS.Name
	newDs.Namespace = m.MEngine.Spec.RemoteNamespace

	newDs.Spec = srcDS.Spec
	// overwrite mode
	newDs.Spec.Mode = v1alpha1.SyncModeRestore

	return m.remoteClient.Create(context.TODO(), newDs)
}

func GetDataSync(kmClient client.Client, name, namespace string) (*v1alpha1.DataSync, error) {
	ds := &v1alpha1.DataSync{}
	err := kmClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, ds)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func GetRemoteDataSync(me *MoveEngineAction) (*v1alpha1.DataSync, error) {
	return GetDataSync(me.remoteClient, me.MEngine.Status.DataSync, me.MEngine.Spec.RemoteNamespace)
}
