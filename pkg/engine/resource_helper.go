package engine

import (
	"context"
	"fmt"

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
				return errors.Errorf("Failed to parse API resource %v.. %v", apiResource.Name, err)
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
		fmt.Printf("Skipping %v\n", api.Name)
		return nil
	}

	// check if ns exists or not
	if api.Name == "namespaces" {
		if err := m.parseNamespace(api); err != nil {
			fmt.Printf("Namespace sync failed %v\n", err)
			return err
		}
	}
	if len(m.mov.Spec.Namespace) != 0 && api.Namespaced {
		if err := m.parseResourceList(api); err != nil {
			fmt.Printf("Failed to create resourceList for %v %v.. %v\n", api.Name, api.Group, err)
			return err
		}
	}

	if !api.Namespaced {
		switch api.Kind {
		case "StorageClass":
			if err := m.parseResourceList(api); err != nil {
				fmt.Printf("Failed to create resourceList for %v %v.. %v\n", api.Name, api.Group, err)
			}
		}
	}
	return nil
}

func (m *MoveEngineAction) parseResourceList(api metav1.APIResource) error {
	list, err := m.ListResourcesFromAPI(api)
	if err != nil {
		fmt.Printf("Failed to fetch list for %v %v.. %v\n", api.Name, api.Group, err)
		return err
	}

	for _, l := range list.Items {
		if err := m.parseResource(api, l); err != nil {
			fmt.Printf("Failed to parse resource for %v %v.. %v\n", api.Name, api.Group, err)
			return err
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
		ns = m.mov.Spec.Namespace
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
			fmt.Printf("Failed to parse volumes for %v/%v\n", obj.GetKind(), obj.GetName())
		}
	}

	if !m.ShouldRestore(obj) {
		fmt.Printf("Skipping %v %v\n", api.Name, obj.GetName())
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
		fmt.Printf("Failed to fetch resource for %v.. %v\n", kind, err)
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
