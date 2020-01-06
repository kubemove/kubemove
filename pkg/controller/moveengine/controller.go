package moveengine

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	v1alpha1 "github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"github.com/kubemove/kubemove/pkg/engine"
	"github.com/kubemove/kubemove/pkg/gcp"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	helper "github.com/vmware-tanzu/velero/pkg/discovery"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_moveengine")

// Add creates a new MoveEngine Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMoveEngine{
		client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		discoveryHelper: nil,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("moveengine-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// activate GCP service account
	if err = gcp.AuthServiceAccount(); err != nil {
		return err
	}

	// Watch for changes to primary resource MoveEngine
	err = c.Watch(&source.Kind{Type: &v1alpha1.MoveEngine{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMoveEngine{}

// ReconcileMoveEngine reconciles a MoveEngine object
type ReconcileMoveEngine struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client          client.Client
	scheme          *runtime.Scheme
	discoveryHelper helper.Helper
	log             logr.Logger
}

// Reconcile reads that state of the cluster for a MoveEngine object and makes changes based on the state read
// and what is in the MoveEngine.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMoveEngine) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling MoveEngine")

	// Fetch the MoveEngine instance
	instance := &v1alpha1.MoveEngine{}
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

	//TODO
	if r.discoveryHelper == nil {
		if err = r.setupHelper(); err != nil {
			r.log.Error(err, "Error setting up helper")
			return reconcile.Result{}, err
		}
	}

	if err = engine.ValidateEngine(instance); err != nil {
		//TODO need to update it in status
		r.log.Error(err, "Validation error")
		return reconcile.Result{}, err
	}

	cr, err := cron.ParseStandard(instance.Spec.SyncPeriod)
	if err != nil {
		r.log.Error(err, "Failed to parse syncPeriod")
		return reconcile.Result{}, err
	}

	if yes, nextSyncDelay := getNextDue(cr, *instance, time.Now()); !yes {
		//TODO job will not be executed as per schedule time, delay due to processing
		// schedule is not yet due
		return reconcile.Result{Requeue: true, RequeueAfter: nextSyncDelay}, nil
	}

	//TODO
	me := engine.NewMoveEngineAction(r.log, r.client, r.discoveryHelper)
	err = me.ParseResourceEngine(instance)
	if err != nil {
		r.log.Error(err, "Failed to parse resources")
	}

	dsName, err := me.CreateDataSync()
	if err != nil || len(dsName) == 0 {
		r.log.Error(err, "Failed to create dataSync CR")
		return reconcile.Result{}, err
	}

	err = me.UpdateMoveEngineStatus(err, dsName)
	if err != nil {
		r.log.Error(err, "Failed to update status")
	}

	_, nextSyncDelay := getNextDue(cr, me.MEngine, time.Now())
	return reconcile.Result{Requeue: true, RequeueAfter: nextSyncDelay}, nil
}

func (r *ReconcileMoveEngine) setupHelper() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Errorf("Failed to fetch config.. %v", err)
	}

	dh, err := helper.NewHelper(
		discovery.NewDiscoveryClientForConfigOrDie(cfg),
		logrus.New(),
	)
	if err != nil {
		return errors.Errorf("Failed to create helper %v", err)
	}

	r.discoveryHelper = dh

	go wait.Forever(
		func() {
			if err := dh.Refresh(); err != nil {
				r.log.Error(err, "Error refreshing discovery")
			}
		},
		//TODO set according to sync period
		5*time.Minute,
		//		signals.SetupSignalHandler(),
	)
	return nil
}

func getNextDue(cr cron.Schedule, move v1alpha1.MoveEngine, now time.Time) (bool, time.Duration) {
	var lastSync time.Time
	if move.Status.SyncedTime.IsZero() {
		if !move.Status.LastSyncedTime.IsZero() {
			// engine hasn't been synced
			lastSync = move.Status.LastSyncedTime.Time
		}
	} else {
		lastSync = move.Status.SyncedTime.Time
	}

	nextSync := cr.Next(lastSync)
	return now.After(nextSync), nextSync.Sub(lastSync)
}
