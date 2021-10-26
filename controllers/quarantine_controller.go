/*
Copyright 2021.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/quarantine"
	"github.com/soer3n/incident-operator/internal/utils"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"
)

const quarantineFinalizer = "finalizer.quarantine.ops.soer3n.info"
const quarantineStatusKey = "active"

// QuarantineReconciler reconciles a Quarantine object
type QuarantineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=ops.soer3n.info,resources=quarantines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ops.soer3n.info,resources=quarantines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ops.soer3n.info,resources=quarantines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Quarantine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.0/pkg/reconcile
func (r *QuarantineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("quarantines", req.NamespacedName)
	_ = r.Log.WithValues("quarantinereq", req)

	// fetch app instance
	instance := &v1alpha1.Quarantine{}

	err := r.Get(ctx, req.NamespacedName, instance)

	reqLogger.Info("starting reconcile loop...")

	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Quarantine resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Quarantine resource")
		return ctrl.Result{}, err
	}

	var q *quarantine.Quarantine
	var requeue bool

	factory := util.NewFactory(genericclioptions.NewConfigFlags(false))
	clientset, err := factory.KubernetesClientSet()

	if err != nil {
		reqLogger.Error(err, "error on initialization of kubernetes clientset")
		return ctrl.Result{}, err
	}

	if q, err = quarantine.New(instance, clientset, factory, reqLogger); err != nil {
		reqLogger.Error(err, "error on initialization of quarantine struct")
		return ctrl.Result{}, err
	}

	if requeue, err = r.handleFinalizer(instance, q, reqLogger); err != nil {
		reqLogger.Error(err, "error on handling resource finalizer")
		return r.syncStatus(context.Background(), instance, reqLogger, metav1.ConditionFalse, "finalizer", err.Error())
	}

	if requeue {
		reqLogger.Info("Update resource after modifying finalizer.")
		if err := r.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "error in reconciling")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if q.IsActive() {
		reqLogger.Info("Quarantine already active. Update if needed.")

		if err := q.Update(); err != nil {
			reqLogger.Error(err, "error in reconciling")
			return r.syncStatus(context.Background(), instance, reqLogger, metav1.ConditionFalse, "update", err.Error())
		}

		return ctrl.Result{}, nil
	}

	reqLogger.Info("preparing...")

	if err := q.Prepare(); err != nil {
		reqLogger.Error(err, "error in reconciling")
		return r.syncStatus(context.Background(), instance, reqLogger, metav1.ConditionFalse, "prepare", err.Error())
	}

	reqLogger.Info("starting...")

	if err := q.Start(); err != nil {
		reqLogger.Error(err, "error in reconciling")
		return r.syncStatus(context.Background(), instance, reqLogger, metav1.ConditionFalse, "starting", err.Error())
	}

	return r.syncStatus(context.Background(), instance, reqLogger, metav1.ConditionTrue, "running", "success")
}

func (r *QuarantineReconciler) handleFinalizer(instance *v1alpha1.Quarantine, obj *quarantine.Quarantine, reqLogger logr.Logger) (bool, error) {

	isResourceMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isResourceMarkedToBeDeleted {
		if err := obj.Stop(); err != nil {
			return true, err
		}

		controllerutil.RemoveFinalizer(instance, quarantineFinalizer)

		return true, nil
	}

	if !utils.Contains(instance.GetFinalizers(), quarantineFinalizer) {
		reqLogger.Info("Adding Finalizer for the Quarantine Resource")
		controllerutil.AddFinalizer(instance, quarantineFinalizer)
		return true, nil
	}

	return false, nil
}

func (r *QuarantineReconciler) syncStatus(ctx context.Context, instance *v1alpha1.Quarantine, reqLogger logr.Logger, stats metav1.ConditionStatus, reason, message string) (ctrl.Result, error) {

	if meta.IsStatusConditionPresentAndEqual(instance.Status.Conditions, quarantineStatusKey, stats) && instance.Status.Conditions[0].Message == message {
		reqLogger.Info("Don't reconcile quarantine resource after sync.")
		return ctrl.Result{}, nil
	}

	condition := metav1.Condition{Type: quarantineStatusKey, Status: stats, LastTransitionTime: metav1.Time{Time: time.Now()}, Reason: reason, Message: message}
	meta.SetStatusCondition(&instance.Status.Conditions, condition)

	if err := r.Status().Update(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	reqLogger.Info("reconcile quarantine resource after status sync.")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuarantineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Quarantine{}).
		WithOptions(controller.Options{CacheSyncTimeout: time.Second * 20}).
		Complete(r)
}

func (r *QuarantineReconciler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.Quarantine{}).
		Complete()
}
