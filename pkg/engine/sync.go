package engine

import (
	"strings"

	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
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
			m.log.Info("Skipping already synced object",
				"APIVersion", v.GetAPIVersion(),
				"Kind", v.GetKind(),
				"Name", v.GetName())
		}

		if err := m.transformObj(v); err != nil {
			m.log.Error(err, "Failed to transform object",
				"APIVersion", v.GetAPIVersion(),
				"Kind", v.GetKind(),
				"Name", v.GetName())
			continue
		}

		if err := m.syncObj(v); err != nil {
			m.log.Error(err, "Failed to create object ",
				"APIVersion", v.GetAPIVersion(),
				"Kind", v.GetKind(),
				"Name", v.GetName())
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
		if len(m.MEngine.Spec.RemoteNamespace) != 0 {
			obj.SetNamespace(m.MEngine.Spec.RemoteNamespace)
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
		//TODO
		unstructured.RemoveNestedField(obj.Object, "metadata", "annotations")
		unstructured.RemoveNestedField(obj.Object, "spec", "volumeName")
	case "PersistentVolume":
		//TODO
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
		gv := gvr.GroupVersion()
		obj.SetAPIVersion(gv.String())
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
		//TODO check if spec is same or not
		if !k8serror.IsAlreadyExists(err) {
			return err
		}
	} else {
		m.log.Info("Object created",
			"Resource", obj.GetAPIVersion(),
			"Kind", obj.GetKind(),
			"Name", obj.GetName())
	}

	m.updateSyncStatus(obj)
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
				m.log.Info("Object is already created",
					"OwnerKind", o.Kind,
					"OwnerName", o.Name,
					"Kind", obj.GetKind(),
					"Name", obj.GetName())
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
					m.log.Error(err, "Failed to fetch object",
						"Namespace", obj.GetNamespace(),
						"Kind", o.Kind,
						"Name", o.Name)
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
	case "StorageClass":
		shouldRestore = true
	}

	return shouldRestore
}
