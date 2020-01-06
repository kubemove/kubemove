package engine

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *MoveEngineAction) UpdateSyncResourceList() error {
	apiResourceList := m.discoveryHelper.Resources()

	//TODO reset seen for multi group Resource
	for _, rg := range apiResourceList {
		gv, err := schema.ParseGroupVersion(rg.GroupVersion)
		if err != nil {
			return errors.Wrapf(err, "Failed to parse GroupVersion %s", rg.GroupVersion)
		}
		for _, apiResource := range rg.APIResources {
			apiResource.Group = gv.Group
			apiResource.Version = gv.Version

			k, ok := m.multiAPIResources[apiResource.Name]
			if ok {
				if k.seen {
					continue
				}
				k.seen = true
			}
			err := m.parseAPIResource(apiResource)
			if err != nil {
				return errors.Wrapf(err, "Failed to parse API resource %v", apiResource.Name)
			}
		}
	}
	return nil
}

func (m *MoveEngineAction) parseAPIResource(api metav1.APIResource) error {
	//TODO move it to fn
	//TODO pass top groupVersion
	switch api.Name {
	case "leases", "nodes", "events":
		m.log.Info("Skipping", "Resource", api.Name)
		return nil
	}

	// check if ns exists or not
	if api.Name == "namespaces" {
		if err := m.parseNamespace(api); err != nil {
			return errors.Wrapf(err, "Namespace sync failed")
		}
	}
	if len(m.MEngine.Spec.Namespace) != 0 && api.Namespaced {
		if err := m.parseResourceList(api); err != nil {
			return errors.Wrapf(err, "Failed to create resourceList for %v %v\n", api.Name, api.Group)
		}
	}

	if !api.Namespaced {
		switch api.Kind {
		case "StorageClass":
			if err := m.parseResourceList(api); err != nil {
				//TODO
				m.log.Error(err, "Failed to create resourceList", "Resource", api.Name, "Group", api.Group)
			}
		}
	}
	return nil
}

func (m *MoveEngineAction) parseResourceList(api metav1.APIResource) error {
	list, err := m.ListResourcesFromAPI(api)
	if err != nil {
		return errors.Wrapf(err, "Failed to fetch list for %v %v\n", api.Name, api.Group)
	}

	for _, l := range list.Items {
		if err := m.parseResource(api, l); err != nil {
			return errors.Wrapf(err, "Failed to parse resource for %v %v\n", api.Name, api.Group)
		}
	}
	return nil
}

func (m *MoveEngineAction) ListResourcesFromAPI(api metav1.APIResource) (*unstructured.UnstructuredList, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   api.Group,
		Version: api.Version,
		Kind:    api.Kind,
	})

	ns := ""
	if api.Namespaced {
		ns = m.MEngine.Spec.Namespace
	}

	err := m.client.List(
		context.TODO(),
		&client.ListOptions{
			Namespace:     ns,
			LabelSelector: m.selector,
		},
		list)
	return list, err

}

func (m *MoveEngineAction) parseResource(api metav1.APIResource, obj unstructured.Unstructured) error {
	switch api.Name {
	case "pods":
		if err := m.parseVolumes(api, obj, m.checkIfSTSPod(obj)); err != nil {
			m.log.Error(err, "Failed to parse volumes for pod", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		}
	}

	if !m.ShouldRestore(obj) {
		m.log.Info("Skipping", "Resource", api.Name, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return nil
	}

	m.addToResourceList(obj)
	return nil
}

func (m *MoveEngineAction) getResource(name, ns, kind string) (unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}

	// TODO check ResourceFor : if returns top one
	gvr, _, err := m.discoveryHelper.ResourceFor(schema.ParseGroupResource(kind).WithVersion(""))
	if err != nil {
		m.log.Error(err, "Failed to fetch resource", "Resource", kind)
		return *obj, nil
	}

	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    kind,
	})
	err = m.client.Get(
		context.TODO(),
		client.ObjectKey{Name: name, Namespace: ns},
		obj)
	return *obj, err
}

func (m *MoveEngineAction) getRemoteResource(name, ns, kind string) (unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}

	gvr, err := m.remoteMapper.ResourceFor(schema.ParseGroupResource(kind).WithVersion(""))
	if err != nil {
		return *obj, err
	}

	gv := gvr.GroupVersion()
	obj.SetAPIVersion(gv.String())
	obj.SetKind(kind)
	err = m.remoteClient.Get(
		context.TODO(),
		client.ObjectKey{
			Name:      name,
			Namespace: ns,
		},
		obj)
	return *obj, err
}

func (m *MoveEngineAction) getResourceFromAPI(api metav1.APIResource, key client.ObjectKey) (unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   api.Group,
		Version: api.Version,
		Kind:    api.Kind,
	})

	err := m.client.Get(
		context.TODO(),
		key,
		obj)
	return *obj, err

}
