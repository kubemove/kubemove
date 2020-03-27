package moveengine

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	errUtil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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

	me, err := engine.NewMoveEngineAction(r.log, r.client, r.discoveryHelper, instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	if EnableResources {
		err = me.ParseResourceEngine(instance)
		if err != nil {
			r.log.Error(err, "Failed to parse resources")
		}
	}

	switch instance.Status.Status {
	case "":
		// We haven't invoked the INIT API yet. Let's initialize the MoveEngine.
		err := r.initializeMoveEngine(me)
		if err != nil {
			return reconcile.Result{}, err
		}
		// MoveEngine has been initialized. Rest of the process should execute on next reconciliation.
		return reconcile.Result{Requeue: true}, nil

	case v1alpha1.MoveEngineInitialized:
		requeue, requeueAfter, err := r.ensureMoveEngineReady(me)
		if err != nil {
			return reconcile.Result{}, err
		}
		if requeue {
			if requeueAfter != nil {
				return reconcile.Result{Requeue: requeue, RequeueAfter: *requeueAfter}, nil
			}
			return reconcile.Result{Requeue: true}, nil
		}

	case v1alpha1.MoveEngineInitializationFailed:
		// MoveEngine has failed to be initialized. There is no point of processing further.
		r.log.Info("Skipping processing MoveEngine: %s/%s. Reason: MoveEngine has failed to be initialized.", instance.Namespace, instance.Name)
		return reconcile.Result{}, nil

	case v1alpha1.MoveEngineReady:
		// MoveEngine is ready. Now, we can start rest of the sync process.
		requeue, requeueAfter, err := r.handleSyncProcess(me)
		if err != nil {
			return reconcile.Result{}, err
		}
		if requeue {
			if requeueAfter != nil {
				return reconcile.Result{Requeue: true, RequeueAfter: *requeueAfter}, nil
			}
			return reconcile.Result{Requeue: true}, nil
		}

	case v1alpha1.MoveEngineInitializing:
		r.log.Info("Skipping processing MoveEngine: %s/%s. Reason: MoveEngine is initializing.", instance.Namespace, instance.Name)
		return reconcile.Result{Requeue: true, RequeueAfter: RetryInterval}, nil

	default:
		r.log.Info(fmt.Sprintf("Status: %s is invalid for MoveEngine.", instance.Status.Status))
		return reconcile.Result{}, fmt.Errorf("invalid MoveEngine Status")
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileMoveEngine) initializeMoveEngine(me *engine.MoveEngineAction) error {
	// Set MoveEngine status to "Initializing"
	err := me.SetMoveEngineState(v1alpha1.MoveEngineInitializing)
	if err != nil {
		return err
	}

	// If MoveEngine mode is active, create a standby MoveEngine in the destination cluster.
	if me.MEngine.Spec.Mode == engine.MoveEngineActive {
		err := me.CreateStandbyMoveEngine()
		if err != nil {
			sErr := me.SetMoveEngineState(v1alpha1.MoveEngineInitializationFailed)
			return errUtil.NewAggregate([]error{err, sErr})
		}
	}

	// Invoke the INIT API
	err = r.ddm.Init(me.MEngine.Spec.PluginProvider, me.MEngine)
	if err != nil {
		r.log.Error(err, "Failed to initialize plugin")
		sErr := me.SetMoveEngineState(v1alpha1.MoveEngineInitializationFailed)
		return errUtil.NewAggregate([]error{err, sErr})
	}

	// MoveEngine has been initialized successfully. So, set MoveEngine status to "Initialized"
	err = me.SetMoveEngineState(v1alpha1.MoveEngineInitialized)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileMoveEngine) handleSyncProcess(me *engine.MoveEngineAction) (bool, *time.Duration, error) {
	requeueAfter := RetryInterval
	// If the MoveEngine is in active mode and no previous sync has done yet or the previous
	// sync has been completed, then start a new sync at a scheduled period.
	if me.MEngine.Spec.Mode == engine.MoveEngineActive &&
		(me.MEngine.Status.DataSync == "" ||
			me.MEngine.Status.DataSyncStatus == v1alpha1.SyncPhaseSynced ||
			me.MEngine.Status.DataSyncStatus == v1alpha1.SyncPhaseFailed) {
		return r.ensureNextSync(me)
	}

	// If the MoveEngine is in standby mode and no previous DataSync has been created or previous DataSync has completed,
	// we nothing to do. When source MoveEngine controller will create a new DataSync, it will also set the
	// respective DataSync name to the standby MoveEngine.
	if me.MEngine.Spec.Mode == engine.MoveEngineStandby && me.MEngine.Status.DataSync == "" {
		return false, nil, nil
	}

	// In active cluster, we should make the DataSyncStatus "Synced" if both of the active and standby MoveEngines has DataSyncStatus "Completed"
	if me.MEngine.Spec.Mode == engine.MoveEngineActive && me.MEngine.Status.DataSyncStatus == v1alpha1.SyncPhaseCompleted {
		// If DataSync CR hasn't been crated in the remote cluster, create one.
		_, err := engine.GetRemoteDataSync(me)
		if err != nil {
			if apierr.IsNotFound(err) {
				return true, &requeueAfter, me.CreateDataSyncAtRemote(me.MEngine.Status.DataSync)
			}
			return false, nil, err
		}
		// DataSync CR is already present in the remote. We should check the sync status.
		status, err := me.GetStandbyMoveEngineStatus()
		if err != nil {
			return false, nil, err
		}
		if status.DataSyncStatus == v1alpha1.SyncPhaseCompleted {
			err := me.UpdateMoveEngineStatus(
				&v1alpha1.MoveEngineStatus{
					DataSyncStatus: v1alpha1.SyncPhaseSynced,
					SyncedTime:     &metav1.Time{Time: time.Now()},
				})
			if err != nil {
				return false, nil, err
			}
			// Also update the standby DataSync CR
			return false, nil, me.UpdateStandbyMoveEngineStatus(
				&v1alpha1.MoveEngineStatus{
					DataSyncStatus: v1alpha1.SyncPhaseSynced,
					SyncedTime:     me.MEngine.Status.SyncedTime,
				})
		}
		return true, &requeueAfter, nil
	}

	// If a sync was running, update the DataSyncStatus status to match with the latest DataSync CR status
	if me.MEngine.Status.DataSync != "" && me.MEngine.Status.DataSyncStatus != v1alpha1.SyncPhaseSynced {
		requeue, err := me.SyncDataSyncStatus()
		if err != nil {
			return false, nil, err
		}
		if requeue {
			return true, &requeueAfter, nil
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (r *ReconcileMoveEngine) ensureNextSync(me *engine.MoveEngineAction) (bool, *time.Duration, error) {
	cr, err := cron.ParseStandard(me.MEngine.Spec.SyncPeriod)
	if err != nil {
		r.log.Error(err, "Failed to parse syncPeriod")
		return false, nil, err
	}

	// If its not time for a scheduled sync yet, reconcile at next sync period
	if yes, nextSyncDelay := getNextDue(cr, me.MEngine, time.Now()); !yes {
		//TODO job will not be executed as per schedule time, delay due to processing
		// schedule is not yet due
		return true, &nextSyncDelay, nil
	}

	// A schedule has appeared. Trigger a sync by creating a DataSync CR.
	dsName, err := me.CreateDataSync()
	if err != nil || len(dsName) == 0 {
		r.log.Error(err, "Failed to create dataSync CR")
		return false, nil, err
	}
	// Add the newly created DataSync CR name in the MoveEngine status.
	err = me.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{DataSync: dsName, DataSyncStatus: v1alpha1.SyncPhaseRunning})
	if err != nil {
		r.log.Error(err, "Failed to update status")
		return false, nil, err
	}

	// Also, add the newly created DataSync CR name in the standby MoveEngine status.
	err = me.UpdateStandbyMoveEngineStatus(&v1alpha1.MoveEngineStatus{DataSync: dsName, DataSyncStatus: v1alpha1.SyncPhaseRunning})
	if err != nil {
		r.log.Error(err, "Failed to update status of standby MoveEngine")
		return false, nil, err
	}

	// Current scheduled sync has been triggered. We have nothing to do until next schedule.
	// So, reconcile at next sync period.
	_, nextSyncDelay := getNextDue(cr, me.MEngine, time.Now())
	return true, &nextSyncDelay, nil
}

func (r *ReconcileMoveEngine) ensureMoveEngineReady(me *engine.MoveEngineAction) (bool, *time.Duration, error) {
	requeueAfter := RetryInterval
	// If the current MoveEngine is in active mode, we need to know the status of standby MoveEngine to determine the final state.
	if me.MEngine.Spec.Mode == engine.MoveEngineActive {
		status, err := me.GetStandbyMoveEngineStatus()
		if err != nil {
			r.log.Error(err, "Failed to determine the status of standby MoveEngine. Reason: %v", err.Error())
			return false, nil, err
		}
		if status.Status == v1alpha1.MoveEngineReady {
			// Standby MoveEngine has been initialized successfully. So, set MoveEngine status to "Ready"
			return false, nil, me.SetMoveEngineState(v1alpha1.MoveEngineReady)
		}

		// Standby MoveEngine is yet to be ready. Retry after a moment.
		return true, &requeueAfter, nil
	} else {
		// In destination cluster, we don't need to check the status of the active MoveEngine.
		// We can just upgrade the status to "Ready".
		return false, nil, me.SetMoveEngineState(v1alpha1.MoveEngineReady)
	}
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
