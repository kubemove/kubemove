package engine

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *MoveEngineAction) parseNamespace(api metav1.APIResource) error {
	if len(m.MEngine.Spec.Namespace) != 0 &&
		len(m.MEngine.Spec.RemoteNamespace) != 0 {
		obj, err := m.getResourceFromAPI(api, client.ObjectKey{Name: m.MEngine.Spec.Namespace})
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch namespace %v", m.MEngine.Spec.Namespace)
		}
		obj.SetName(m.MEngine.Spec.RemoteNamespace)
		m.addToResourceList(obj)
	} else {
		list, err := m.ListResourcesFromAPI(api)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch namespace list")
		}
		for _, i := range list.Items {
			m.addToResourceList(i)
		}
	}
	return nil
}
