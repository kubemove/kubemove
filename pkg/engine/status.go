package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//TODO
// new package for status

//TODO
var PVCWaitTime = 5 * time.Second
var PVCWaitInterval = 1 * time.Second

type resourceStatusFn func(*MoveEngineAction, unstructured.Unstructured) interface{}

var statusAction map[string]resourceStatusFn

func init() {
	//TODO move to global action i.e postSync Action
	statusAction = map[string]resourceStatusFn{
		"deployment":            deploymentStatus,
		"persistentvolume":      pvStatus,
		"persistentvolumeclaim": pvcStatus,
		"namespace":             nsStatus,
	}
}

func (m *MoveEngineAction) updateSyncStatus(obj unstructured.Unstructured) {
	var (
		rs   *v1alpha1.ResourceStatus
		pvRs *v1alpha1.VolumeStatus
		ok   bool
	)

	rs = newResourceStatus(obj)
	rs.Phase = "Synced"

	kind := strings.ToLower(obj.GetKind())
	fn, ok := statusAction[kind]
	if ok {
		newRs := fn(m, obj)
		if newRs == nil {
			m.log.Error(nil, "Unable to update resourceStatus", "Resource", obj.GetKind(), "Name", obj.GetName())
		} else {
			// If it is PV then we will skip the syncedResourceList and add it to syncedVolMap
			switch kind {
			case "persistentvolume":
				pvRs, ok = newRs.(*v1alpha1.VolumeStatus)
				if ok {
					m.addToSyncedVolumesList(obj, *pvRs)
				}
				return
			default:
				var r *v1alpha1.ResourceStatus
				r, ok = newRs.(*v1alpha1.ResourceStatus)
				if ok {
					rs = r
				}
			}
			if !ok {
				m.log.Error(nil, fmt.Sprintf("Failed to parse status.. unexpected type %T", newRs), "Resource", kind, "Name", obj.GetName())
			}
		}
	}

	if kind != "persistentvolume" {
		m.addToSyncedResourceList(obj, *rs)
	}
	return
}

func deploymentStatus(m *MoveEngineAction, obj unstructured.Unstructured) interface{} {
	deploy := new(appsv1.Deployment)
	rs := newResourceStatus(obj)
	rs.Phase = "Synced"

	newObj, err := fetchRemoteResourceFromObj(m.remoteClient, obj)
	if err == nil {
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(newObj.UnstructuredContent(), deploy); err == nil {
			for _, d := range deploy.Status.Conditions {
				rs.Status = string(d.Type)
				rs.Reason = d.Reason
				break
			}
		} else {
			rs.Reason = err.Error()
		}
	} else {
		rs.Reason = err.Error()
	}
	return rs
}

func pvStatus(m *MoveEngineAction, obj unstructured.Unstructured) interface{} {
	pv := new(v1.PersistentVolume)
	vs := newVolumeStatus(obj)

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), pv); err == nil {
		vs.Status = string(pv.Status.Phase)

		if yes := isPVDynamicallyProvisioned(obj); yes {
			if pv.Spec.ClaimRef != nil {
				claimRef := pv.Spec.ClaimRef
				switch claimRef.Kind {
				case "PersistentVolumeClaim":
					vs.PVC = claimRef.Name
					//TODO Need to fix.
					if len(m.MEngine.Spec.Namespace) != 0 {
						vs.Namespace = m.MEngine.Spec.Namespace
					} else {
						//TODO what if namespace re-mapping is used?
						vs.Namespace = claimRef.Namespace
					}
					vs.RemoteNamespace = claimRef.Namespace
					pvcObj, err := m.getResource(vs.PVC, vs.Namespace, "PersistentVolumeClaim")
					if err == nil {
						pvName, ok, err := unstructured.NestedString(pvcObj.Object, "spec", "volumeName")
						if !ok || err != nil {
							m.log.Error(nil, "No volume mentioned in PVC", "PVC", vs.PVC)
						} else {
							vs.Volume = pvName
						}
					} else {
						m.log.Error(nil, fmt.Sprintf("PVC %v is not added to resource list", vs.PVC))
					}
				}
			}
		} else {
			vs.Volume = obj.GetName()
		}
	} else {
		vs.Reason = err.Error()
	}
	return vs
}

func pvcStatus(m *MoveEngineAction, obj unstructured.Unstructured) interface{} {
	pvc := new(v1.PersistentVolumeClaim)

	rs := newResourceStatus(obj)
	rs.Phase = "Synced"

	_ = wait.PollImmediate(PVCWaitInterval, PVCWaitTime, func() (bool, error) {
		newObj, err := fetchRemoteResourceFromObj(m.remoteClient, obj)
		if err == nil {
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(newObj.UnstructuredContent(), pvc); err == nil {
				rs.Status = string(pvc.Status.Phase)
				for _, c := range pvc.Status.Conditions {
					rs.Reason = c.Reason
					break
				}

				if pvc.Status.Phase == v1.ClaimBound {
					pvName, ok, err := unstructured.NestedString(newObj.Object, "spec", "volumeName")
					if !ok || err != nil {
						m.log.Error(err, "Failed to parse volume from PVC", "Namespace", newObj.GetNamespace(), "Name", newObj.GetName())
						//TODO should we report error?
					} else {
						pvObj, err := m.getRemoteResource(pvName, "", "PersistentVolume")
						if err != nil {
							m.log.Error(err, "Failed to fetch PV", "PV", pvName)
							//TODO should we report error?
						} else {
							m.updateSyncStatus(pvObj)
							return true, nil
						}
					}
				} else if pvc.Status.Phase == v1.ClaimLost {
					return true, nil
				}
			} else {
				rs.Reason = err.Error()
			}
		} else {
			rs.Reason = err.Error()
		}
		return false, nil
	})

	return rs
}

func nsStatus(m *MoveEngineAction, obj unstructured.Unstructured) interface{} {
	ns := new(v1.Namespace)
	rs := newResourceStatus(obj)
	rs.Phase = "Synced"

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), ns); err == nil {
		rs.Status = string(ns.Status.Phase)
	} else {
		rs.Reason = err.Error()
	}

	return rs
}

func newResourceStatus(obj unstructured.Unstructured) *v1alpha1.ResourceStatus {
	// TODO need to add clock at upper level, moveengineaction
	timestamp := time.Now()
	return &v1alpha1.ResourceStatus{
		Kind:       obj.GetKind(),
		Name:       obj.GetName(),
		SyncedTime: metav1.Time{Time: timestamp},
	}
}

func newVolumeStatus(obj unstructured.Unstructured) *v1alpha1.VolumeStatus {
	// TODO need to add clock at upper level, moveengineaction
	timestamp := time.Now()
	return &v1alpha1.VolumeStatus{
		RemoteVolume: obj.GetName(),
		SyncedTime:   metav1.Time{Time: timestamp},
	}
}

func fetchRemoteResourceFromObj(cl client.Client, obj unstructured.Unstructured) (*unstructured.Unstructured, error) {
	newObj := &unstructured.Unstructured{}
	newObj.SetAPIVersion(obj.GetAPIVersion())
	newObj.SetKind(obj.GetKind())
	err := cl.Get(
		context.TODO(),
		client.ObjectKey{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		},
		newObj)
	return newObj, err
}
