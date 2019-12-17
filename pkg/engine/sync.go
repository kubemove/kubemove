package engine

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var topresource = []string{
	"customresourcedefinition",
	"namespace",
	"storageclass",
	"serviceaccount",
	"customresourcedefinition",
	"secret",
	"configmap",
	"persistentvolume",
	"persistentvolumeclaim",
	"limitrange",
	"statefulset",
	"deployment",
	"daemonset",
	"replicaset",
	"pod",
}

func (m *MoveEngineAction) CreateResourceAtRemote(gvr schema.GroupVersionResource) error {
	objMap, ok := m.resourcesMap[gvr]
	if !ok {
		return nil
	}

	for _, v := range objMap {
		if m.isResourceSynced(v) {
			fmt.Printf("Skipping synced object %v/%v/%v, since already synced\n", v.GetAPIVersion(), v.GetKind(), v.GetName())
			continue
		}

		if err := m.transformObj(v); err != nil {
			fmt.Printf("Failed to transform object %v/%v/%v.. %v\n", v.GetAPIVersion(), v.GetKind(), v.GetName(), err)
			continue
		}

		if err := m.syncObj(v); err != nil {
			fmt.Printf("Failed to create object %v/%v/%v.. %v\n", v.GetAPIVersion(), v.GetKind(), v.GetName(), err)
			continue
		}
	}
	return nil
}

func (m *MoveEngineAction) syncHighPrioResources() error {
	for _, r := range topresource {
		gvr, err := m.remoteMapper.ResourceFor(schema.ParseGroupResource(r).WithVersion(""))
		if err != nil {
			return err
		}
		k, err := m.remoteMapper.KindFor(gvr)
		if err != nil {
			return err
		}
		gvr.Resource = strings.ToLower(k.Kind)
		if err = m.CreateResourceAtRemote(gvr); err != nil {
			return err
		}
	}
	return nil
}

func (m *MoveEngineAction) syncResourceList() error {
	err := m.syncHighPrioResources()
	if err != nil {
		return err
	}

	apiResourceList := m.discoveryHelper.Resources()

	//TODO reset seen for multi group Resource
	for _, rg := range apiResourceList {
		gv, err := schema.ParseGroupVersion(rg.GroupVersion)
		if err != nil {
			return errors.Wrapf(err, "Failed to parse GroupVersion %s", rg.GroupVersion)
		}
		for _, apiResource := range rg.APIResources {
			k, ok := m.multiAPIResources[apiResource.Name]
			if ok {
				if k.seen {
					continue
				}
				k.seen = true
			}
			err := m.CreateResourceAtRemote(
				schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: strings.ToLower(apiResource.Kind),
				},
			)
			if err != nil {
				return errors.Errorf("Failed to create resource.. %v\n", err)
			}
		}
	}
	return nil
}

func (m *MoveEngineAction) transformObj(obj unstructured.Unstructured) error {
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "selfLink")
	unstructured.RemoveNestedField(obj.Object, "spec", "status")

	if len(obj.GetNamespace()) != 0 {
		if len(m.mov.Spec.RemoteNamespace) != 0 {
			obj.SetNamespace(m.mov.Spec.RemoteNamespace)
		}
	}

	switch obj.GetKind() {
	case "Service":
		//TODO Handle this separately
		val, found, err := unstructured.NestedString(obj.Object, "spec", "clusterIP")
		if err == nil && found && val != "None" {
			_ = unstructured.SetNestedField(obj.Object, "", "spec", "clusterIP")
		}
	case "PersistentVolumeClaim":
		unstructured.RemoveNestedField(obj.Object, "metadata", "annotations")
		unstructured.RemoveNestedField(obj.Object, "spec", "volumeName")
	case "PersistentVolume":
		unstructured.RemoveNestedField(obj.Object, "metadata", "annotations")
		unstructured.RemoveNestedField(obj.Object, "spec", "claimRef")
	}
	return nil
}

func (m *MoveEngineAction) syncObj(obj unstructured.Unstructured) error {
	//TODO check
	var err error
	var gvr schema.GroupVersionResource

	gvrlist, err := m.remoteMapper.ResourcesFor(schema.ParseGroupResource(obj.GetKind()).WithVersion(""))
	if err == nil && len(gvrlist) != 0 {
		gvr = gvrlist[0]
	} else {
		gv, _ := schema.ParseGroupVersion(obj.GetAPIVersion())
		gvr = schema.GroupVersionResource{
			Group:    gv.Group,
			Version:  gv.Version,
			Resource: obj.GetKind(),
		}
	}

	_, err = m.remotedyClient.
		Resource(gvr).
		Namespace(obj.GetNamespace()).
		Create(&obj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Object created %v/%v/%v\n", obj.GetAPIVersion(), obj.GetKind(), obj.GetName())
	m.addToSyncedResourceList(obj)
	//TODO if volume or PVC need to append in list
	return nil
}

func (m *MoveEngineAction) ShouldRestore(obj unstructured.Unstructured) bool {
	shouldRestore := true

	switch obj.GetKind() {
	case "Pod", "ReplicaSet":
		//TODO
		// This will be removed by OwnerReference check
		or := obj.GetOwnerReferences()
		for _, o := range or {
			sr := newMResourceFromOR(o)
			if _, ok := m.getFromResourceList(sr); ok {
				fmt.Printf("%v/%v already created by %v/%v\n", obj.GetKind(), obj.GetName(), o.Kind, o.Name)
				shouldRestore = false
				break
			} else if o.Kind == "ReplicaSet" {
				shouldRestore = m.ShouldRestoreRS(o.Name, obj.GetNamespace())
			} else {
				// STS or deployment is not added to syncResource list
				// Let's add it
				ro, err := m.getResource(o.Name, obj.GetNamespace(), o.Kind)
				if err != nil {
					//TODO check if not exist
					fmt.Printf("Failed to fetch %v/%v/%v.. %v\n", obj.GetNamespace(), o.Kind, o.Name, err)
					continue
				}
				//TODO refactor
				sr := newMResourceFromObj(ro)
				if _, ok := m.getFromResourceList(sr); !ok {
					m.addToResourceList(ro)
				}
				shouldRestore = false
			}
		}
	case "Node", "Event":
		shouldRestore = false
	case "Service", "Endpoints":
		m.addToExposedResourceList(obj)
		//TODO check if switch call
		shouldRestore = false
	}

	return shouldRestore
}
