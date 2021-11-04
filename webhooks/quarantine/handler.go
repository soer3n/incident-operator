package quarantine

import (
	"context"
	"errors"
	"net/http"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var err error
var pod *corev1.Pod

const quarantineControllerLabelKey = "component"
const quarantineControllerLabelValue = "incident-controller-manager"

// Handle handles admission requests.
func (h *QuarantineHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	if err := h.manageObject(req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch t := req.Operation; t {
	case admissionv1.Create:
		if err := h.ValidateCreate(); err != nil {
			return admission.Denied(err.Error())
		}
	case admissionv1.Update:
		if err := h.ValidateUpdate(req.OldObject.Object); err != nil {
			return admission.Denied(err.Error())
		}
	case admissionv1.Delete:
		if err := h.ValidateDelete(); err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("controller is on a valid node")
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineHandler) ValidateCreate() error {
	h.Log.Info("validate create", "name", h.Object.Name)

	if pod, err = h.getControllerPod(); err != nil {
		h.Log.Info("error on getting controller pod")
		return err
	}

	if h.controllerShouldBeRescheduled(pod.Spec.NodeName) {
		h.Log.Info("controller pod is on a node marked for isolation")
		return err
	}

	h.Log.Info("controller pod is on a valid node")

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineHandler) ValidateUpdate(old runtime.Object) error {
	h.Log.Info("validate update", "name", h.Object.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineHandler) ValidateDelete() error {
	h.Log.Info("validate delete", "name", h.Object.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (h *QuarantineHandler) controllerShouldBeRescheduled(nodeName string) bool {

	for _, n := range h.Object.Spec.Nodes {
		if n.Name == nodeName {
			return true
		}
	}

	return false
}

func (h *QuarantineHandler) manageObject(req admission.Request) error {

	quarantine := &v1alpha1.Quarantine{}

	if h.Object == nil {
		if err := h.Decoder.Decode(req, quarantine); err != nil {
			return err
		}

	}

	h.Object = quarantine
	return nil
}

func (h *QuarantineHandler) getControllerPod() (*corev1.Pod, error) {

	pods := &corev1.PodList{}
	var pod *corev1.Pod
	var err error

	listOpts := client.ListOptions{
		Raw: &metav1.ListOptions{
			LabelSelector: quarantineControllerLabelKey + "=" + quarantineControllerLabelValue,
		},
	}

	if err = h.Client.List(context.TODO(), pods, &listOpts); err != nil {
		return pod, err
	}

	if len(pods.Items) > 1 {
		return pod, errors.New("multiple controller pods found")
	}

	return &pods.Items[0], nil
}
