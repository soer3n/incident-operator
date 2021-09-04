/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	opsv1alpha1 "github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/quarantine"
	"github.com/soer3n/incident-operator/internal/utils"
)

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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *QuarantineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("repos", req.NamespacedName)
	_ = r.Log.WithValues("reposreq", req)

	// fetch app instance
	instance := &opsv1alpha1.Quarantine{}

	err := r.Get(ctx, req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Quarantine resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Quarantine resource")
		return ctrl.Result{}, err
	}

	var q *quarantine.Quarantine

	if q, err = quarantine.New(instance); err != nil {
		return ctrl.Result{}, err
	}

	if q.IsActive() {
		reqLogger.Info("Quarantine already active. Updating it if needed.")
		return ctrl.Result{}, q.Update()
	}

	if err := q.Prepare(); err != nil {
		reqLogger.Info("preparing...")
		return ctrl.Result{}, err
	}

	reqLogger.Info("starting...")
	return ctrl.Result{}, q.Start()
}

func (r *QuarantineReconciler) handleFinalizer(instance *v1alpha1.Quarantine, obj *quarantine.Quarantine) error {

	isRepoMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isRepoMarkedToBeDeleted {
		if err := obj.Stop(); err != nil {
			return err
		}

		controllerutil.RemoveFinalizer(instance, "finalizer.quarantine.ops.soer3n.info")

		if err := r.Update(context.Background(), instance); err != nil {
			return err
		}

		return nil
	}

	if utils.Contains(instance.GetFinalizers(), "finalizer.releases.helm.soer3n.info") {
		if err := r.addFinalizer(instance); err != nil {
			return err
		}

		if err := r.Update(context.Background(), instance); err != nil {
			return err
		}
	}

	return nil
}

func (r *QuarantineReconciler) addFinalizer(q *v1alpha1.Quarantine) error {
	log.Info("Adding Finalizer for the Quarantine Resource")
	controllerutil.AddFinalizer(q, "quarantine.ops.soer3n.info")

	// Update CR
	if err := r.Update(context.TODO(), q); err != nil {
		log.Error(err, "Failed to add finalizer to Quarantine resource")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuarantineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1alpha1.Quarantine{}).
		Complete(r)
}
