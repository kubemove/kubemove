package movepair

import (
	"context"

	"github.com/go-logr/logr"
	kubemovev1alpha1 "github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"github.com/kubemove/kubemove/pkg/gcp"
	kmpair "github.com/kubemove/kubemove/pkg/pair"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_movepair")

// Add creates a new MovePair Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMovePair{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("movepair-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// activate GCP service account
	if err = gcp.AuthServiceAccount(); err != nil {
		return err
	}

	// Watch for changes to primary resource MovePair
	err = c.Watch(&source.Kind{Type: &kubemovev1alpha1.MovePair{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMovePair{}

// ReconcileMovePair reconciles a MovePair object
type ReconcileMovePair struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

// Reconcile reads that state of the cluster for a MovePair object and makes changes based on the state read
// and what is in the MovePair.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMovePair) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling MovePair")

	// Fetch the MovePair instance
	instance := &kubemovev1alpha1.MovePair{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	status, err := r.verifyMovePairStatus(instance)
	if err != nil {
		r.log.Error(err, "Failed to verify movePair")
	}

	err = r.updateStatus(instance, status)
	if err != nil {
		r.log.Error(err, "Failed to update movePair status")
		//TODO requeue
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileMovePair) verifyMovePairStatus(mpair *kubemovev1alpha1.MovePair) (string, error) {
	err := clientcmd.Validate(mpair.Spec.Config)
	if err != nil {
		return "Errored", err
	}

	dclient, err := kmpair.FetchPairDiscoveryClient(mpair)
	if err != nil {
		return "Errored", err
	}

	// To verify access, let's fetch remote server version
	_, err = dclient.ServerVersion()
	if err != nil {
		return "Errored", err
	}
	return "Connected", nil
}

// update movePair status
func (r *ReconcileMovePair) updateStatus(mpair *kubemovev1alpha1.MovePair, status string) error {
	mpair.Status.State = status
	return r.client.Update(context.TODO(), mpair)
}
