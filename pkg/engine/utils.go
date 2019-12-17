package engine

import (
	"strings"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (m *MoveEngineAction) addToResourceList(obj unstructured.Unstructured) {
	sr := newMResourceFromObj(obj)
	gv, _ := schema.ParseGroupVersion(sr.APIVersion)
	gvr := schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: sr.Kind,
	}
	o, ok := m.resourcesMap[gvr]
	if !ok {
		nm := make(map[string]unstructured.Unstructured)
		m.resourcesMap[gvr] = nm
		nm[sr.Name] = obj
	} else {
		o[sr.Name] = obj
	}

	//TODO If resource is having multiple version let's add it to other groups also

	m.resourceList = append(m.resourceList, obj)
}

func (m *MoveEngineAction) getFromResourceList(sr MResources) (unstructured.Unstructured, bool) {
	gv, _ := schema.ParseGroupVersion(sr.APIVersion)
	gvr := schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: sr.Kind,
	}
	o, ok := m.resourcesMap[gvr]
	if !ok {
		return unstructured.Unstructured{}, false
	}
	obj, ok := o[sr.Name]
	return obj, ok
}

func (m *MoveEngineAction) addToExposedResourceList(obj unstructured.Unstructured) {
	sr := newMResourceFromObj(obj)
	m.exposedResourceMap[sr] = obj
}

func (m *MoveEngineAction) addToVolumeList(obj unstructured.Unstructured) {
	sr := newMResourceFromObj(obj)
	m.volMap[sr] = obj
	m.addToResourceList(obj)
}

func (m *MoveEngineAction) addToSTSVolumeList(obj unstructured.Unstructured) {
	sr := newMResourceFromObj(obj)
	m.stsVolMap[sr] = obj
}

func (m *MoveEngineAction) addToSyncedResourceList(obj unstructured.Unstructured) {
	sr := newMResourceFromObj(obj)
	//TODO
	m.syncedResourceMap[sr] = v1alpha1.ResourceStatus{}
}

func (m *MoveEngineAction) isResourceSynced(obj unstructured.Unstructured) bool {
	sr := newMResourceFromObj(obj)
	if _, ok := m.syncedResourceMap[sr]; ok {
		return true
	}
	return false
}

func (m *MoveEngineAction) addToSyncedVolumesList(obj unstructured.Unstructured) {
	sr := newMResourceFromObj(obj)
	m.syncedVolMap[sr] = v1alpha1.VolumeStatus{}
}

func newMResourceFromObj(obj unstructured.Unstructured) MResources {
	return MResources{
		Name:       obj.GetName(),
		Kind:       strings.ToLower(obj.GetKind()),
		APIVersion: obj.GetAPIVersion(),
	}
}

func newMResourceFromOR(or metav1.OwnerReference) MResources {
	return MResources{
		Name:       or.Name,
		Kind:       strings.ToLower(or.Kind),
		APIVersion: or.APIVersion,
	}
}
