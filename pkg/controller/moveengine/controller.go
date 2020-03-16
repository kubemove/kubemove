package moveengine

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1alpha1 "github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"github.com/kubemove/kubemove/pkg/engine"
	"github.com/kubemove/kubemove/pkg/gcp"
	"github.com/kubemove/kubemove/pkg/plugin/ddm/server"
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

const (
	RetryInterval = 30 * time.Second
)

var log = logf.Log.WithName("controller_moveengine")

// EnableResources defines if resources needs to be migrated
var EnableResources bool

// Add creates a new MoveEngine Controller and adds it to the Manager. The Manager will set fields on the Controller
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

	return &ReconcileMoveEngine{
		client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		discoveryHelper: nil,
		ddm:             ddm,
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
	ddm             server.DDM
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
	if r.discoveryHelper == nil && EnableResources {
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

	//TODO
	me := engine.NewMoveEngineAction(r.log, r.client, r.discoveryHelper)
	if EnableResources {
		err = me.ParseResourceEngine(instance)
		if err != nil {
			r.log.Error(err, "Failed to parse resources")
		}
	}

	switch instance.Status.Status {
	case "": // We haven't invoked the INIT API yet.
		// Set MoveEngine status to "Initializing"
		err := me.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{Status: v1alpha1.MoveEngineInitializing}, "")
		if err != nil {
			return reconcile.Result{}, err
		}

		// Invoke the INIT API
		err = r.ddm.Init(instance.Spec.PluginProvider, *instance)
		if err != nil {
			// If the error is due to timeout, then just requeue after some time. Next step will take care of it.
			if apierr.IsTimeout(err) { // TODO: parse actual timeout message.
				return reconcile.Result{Requeue: true, RequeueAfter: RetryInterval}, nil
			}
			r.log.Error(err, "Failed to initialize plugin")
			return reconcile.Result{}, err
		}

		// MoveEngine has been initialized successfully. So, set MoveEngine status to "Initialized"
		err = me.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{Status: v1alpha1.MoveEngineInitialized}, "")
		if err != nil {
			return reconcile.Result{}, err
		}
	case v1alpha1.MoveEngineInitializing:
		// Invoke INIT_PROGRESS API to determine whether the plugin has completed it's initialization or not.
		state, err := r.ddm.InitProgress(instance.Spec.PluginProvider, *instance)
		if err != nil {
			r.log.Error(err, "Failed to determine initialization progress. Reason: %v", err.Error())
			return reconcile.Result{}, err
		}
		// MoveEngine has been initialized successfully. So, set MoveEngine status to "Initialized"
		err = me.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{Status: v1alpha1.MoveEngineState(state)}, "")
		if err != nil {
			return reconcile.Result{}, err
		}
	case v1alpha1.MoveEngineInitialized:
		// MoveEngine of the source has been initialized. Now, we need to the MoveEngine status of the destination cluster.
		// TODO: determine MoveEngine status of the destination cluster
		status, err := me.GetStandbyEngineStatus()
		if err != nil {
			r.log.Error(err, "Failed to determine the status of standby MoveEngine. Reason: %v", err.Error())
			return reconcile.Result{}, err
		}
		if status == v1alpha1.MoveEngineInitialized {
			// Standby MoveEngine has been initialized successfully. So, set MoveEngine status to "Ready"
			err = me.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{Status: v1alpha1.MoveEngineReady}, "")
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case v1alpha1.MoveEngineInitializationFailed:
		// MoveEngine has failed to initialize. There is no point of processing further.
		r.log.Info("Skipping processing MoveEngine: %s/%s. Reason: MoveEngine has failed to be initialized.", instance.Namespace, instance.Name)
		return reconcile.Result{}, nil
	case v1alpha1.MoveEngineReady:
		// MoveEngine is ready. Now, we can start rest of the sync process
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

		dsName, err := me.CreateDataSync()
		if err != nil || len(dsName) == 0 {
			r.log.Error(err, "Failed to create dataSync CR")
			return reconcile.Result{}, err
		}

		err = me.UpdateMoveEngineStatus(nil, dsName)
		if err != nil {
			r.log.Error(err, "Failed to update status")
		}

		_, nextSyncDelay := getNextDue(cr, me.MEngine, time.Now())
		return reconcile.Result{Requeue: true, RequeueAfter: nextSyncDelay}, nil
	default:
		r.log.Info("Status: %s is invalid for MoveEngine.", instance.Status.Status)
		return reconcile.Result{}, fmt.Errorf("invalid MoveEngine Status")
	}
	return reconcile.Result{}, nil
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
