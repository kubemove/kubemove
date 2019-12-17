package engine

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func (m *MoveEngineAction) checkIfSTSPod(obj unstructured.Unstructured) bool {
	or := obj.GetOwnerReferences()
	for _, o := range or {
		if o.Kind == "StatefulSet" {
			return true
		}
	}
	return false
}

func (m *MoveEngineAction) ShouldRestoreRS(name, ns string) bool {
	//TODO add check from cmd line
	shouldRestore := true
	if obj, err := m.getResource(name, ns, "ReplicaSet"); err == nil {
		shouldRestore = m.ShouldRestore(obj)
		if shouldRestore {
			m.addToResourceList(obj)
		}
		shouldRestore = false
	}
	return shouldRestore
}
