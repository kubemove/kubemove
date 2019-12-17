package engine

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *MoveEngineAction) parseNamespace(api metav1.APIResource) error {
	if len(m.mov.Spec.Namespace) != 0 &&
		len(m.mov.Spec.RemoteNamespace) != 0 {
		obj, err := m.getResourceFromAPI(api, client.ObjectKey{Name: m.mov.Spec.Namespace})
		if err != nil {
			fmt.Printf("Failed to fetch resource %v\n", err)
			return err
		}
		obj.SetName(m.mov.Spec.RemoteNamespace)
		m.addToResourceList(obj)
	} else {
		list, err := m.ListResourcesFromAPI(api)
		if err != nil {
			fmt.Printf("Failed to fetch list for %v %v.. %v\n", api.Name, api.Group, err)
			return err
		}
		for _, i := range list.Items {
			m.addToResourceList(i)
		}
	}
	return nil
}
