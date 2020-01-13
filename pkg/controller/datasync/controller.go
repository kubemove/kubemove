package datasync

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	kubemovev1alpha1 "github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	ds "github.com/kubemove/kubemove/pkg/datasync"
	server "github.com/kubemove/kubemove/pkg/plugin/ddm/server"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	ddm := ds.NewDataSyncer(instance, r.ddm)
	if ddm == nil {
		return reconcile.Result{}, errors.New("Failed to create DataSyncer")
	}

	err = ddm.Sync()
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
