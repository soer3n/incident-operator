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

package v1alpha1

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var quarantinelog = logf.Log.WithName("quarantine-resource")
var k8sClient client.Client

func (r *Quarantine) SetupWebhookWithManager(mgr ctrl.Manager) error {

	k8sClient = mgr.GetClient()

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate,mutating=false,failurePolicy=fail,sideEffects=None,groups=ops.soer3n.info,resources=quarantines,verbs=create;update,versions=v1alpha1,name=vquarantine.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Quarantine{}
var err error
var pod *corev1.Pod

const quarantineControllerLabelKey = "component"
const quarantineControllerLabelValue = "incident-controller-manager"

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Quarantine) ValidateCreate() error {
	quarantinelog.Info("validate create", "name", r.Name)

	if pod, err = r.getControllerPod(); err != nil {
		quarantinelog.Info("error on getting controller pod")
		return errors.New("error on getting controller pod")
	}

	if r.controllerShouldBeRescheduled(pod.Spec.NodeName) {
		quarantinelog.Info("controller pod is on a node marked for isolation")
		return errors.New("controller pod is on a node marked for isolation")
	}

	quarantinelog.Info("controller pod is on a valid node")

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Quarantine) ValidateUpdate(old runtime.Object) error {
	quarantinelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Quarantine) ValidateDelete() error {
	quarantinelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Quarantine) controllerShouldBeRescheduled(nodeName string) bool {

	for _, n := range r.Spec.Nodes {
		if n.Name == nodeName {
			return true
		}
	}

	return false
}

func (r *Quarantine) getControllerPod() (*corev1.Pod, error) {

	var pods *corev1.PodList
	var pod *corev1.Pod
	var err error

	listOpts := client.ListOptions{
		Raw: &metav1.ListOptions{
			LabelSelector: quarantineControllerLabelKey + "=" + quarantineControllerLabelValue,
		},
	}

	if err = k8sClient.List(context.TODO(), pods, &listOpts); err != nil {
		return pod, err
	}

	if len(pods.Items) > 1 {
		return pod, errors.New("multiple controller pods found")
	}

	return &pods.Items[0], nil
}
