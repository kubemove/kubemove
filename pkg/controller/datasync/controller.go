package datasync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	kubemovev1alpha1 "github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	ds "github.com/kubemove/kubemove/pkg/datasync"
	plugin "github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
	server "github.com/kubemove/kubemove/pkg/plugin/ddm/server"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	RetryInterval = 5 * time.Second
)

var log = logf.Log.WithName("controller_datasync")

// Add creates a new DataSync Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ddm, err := server.NewDDMServer()
	if err != nil {
		log.Error(err, "Failed to create DDM server")
		return nil
	}

	return &ReconcileDataSync{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		ddm:    ddm,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("datasync-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DataSync
	err = c.Watch(&source.Kind{Type: &kubemovev1alpha1.DataSync{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDataSync{}

// ReconcileDataSync reconciles a DataSync object
type ReconcileDataSync struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
	ddm    server.DDM
}

// Reconcile reads that state of the cluster for a DataSync object and makes changes based on the state read
// and what is in the DataSync.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDataSync) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling DataSync")

	// Fetch the DataSync instance
	instance := &kubemovev1alpha1.DataSync{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierr.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	ddm := ds.NewDataSyncer(r.client, instance, r.ddm)
	if ddm == nil {
		return reconcile.Result{}, errors.New("failed to create DataSyncer")
	}

	switch instance.Status.Status {
	case "":
		// No sync has been done for this DataSync CR. Let's invoke the SYNC API and set it's status to "Running".
		err := ddm.UpdateDataSyncStatus(kubemovev1alpha1.SyncStateRunning, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		// Invoke the Sync API
		err = ddm.Sync()
		if err != nil {
			return reconcile.Result{}, err
		}
		// Now, we should hit STATUS API periodically to know the sync status.
		// Here, we will utilize the controller to achieve the periodic check behavior.
		// We will just requeue after some time. Rest of the process will be taken care of by the respective case.
		return reconcile.Result{Requeue: true, RequeueAfter: RetryInterval}, nil
	case kubemovev1alpha1.SyncStateRunning:
		// Invoke the STATUS API to determine the sync status
		state, err := ddm.SyncStatus()
		if err != nil && state != plugin.Errored {
			return reconcile.Result{}, err
		}
		switch state {
		case plugin.InProgress:
			// Nothing to do. Just wait and try after sometime
			return reconcile.Result{Requeue: true, RequeueAfter: RetryInterval}, nil
		case plugin.Completed:
			// Set DataSync status to completed
			return reconcile.Result{}, ddm.UpdateDataSyncStatus(kubemovev1alpha1.SyncStateCompleted, "")
		case plugin.Errored, plugin.Unknown:
			// Set DataSync status to failed and set the respective error
			failureReason := ""
			if err != nil {
				failureReason = err.Error()
			}
			return reconcile.Result{}, ddm.UpdateDataSyncStatus(kubemovev1alpha1.SyncStateFailed, failureReason)
		case plugin.Failed:
			// Sync failed but failure reason is unknown
			return reconcile.Result{}, ddm.UpdateDataSyncStatus(kubemovev1alpha1.SyncStateFailed, "")
		default:
			return reconcile.Result{}, fmt.Errorf("unknown Sync state")
		}
	case kubemovev1alpha1.SyncStateCompleted, kubemovev1alpha1.SyncStateFailed:
		// Nothing to do. Just log a message and skip processing
		r.log.Info(fmt.Sprintf("Skipping processing DataSync %s/%s. Reason: Status is %q", instance.Namespace, instance.Name, instance.Status.Status))
		return reconcile.Result{}, nil
	default:
		return reconcile.Result{}, fmt.Errorf("invalid status: %q for DataSync: %s/%s", instance.Status.Status, instance.Namespace, instance.Name)
	}
}
