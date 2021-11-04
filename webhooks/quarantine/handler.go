package quarantine

import (
	"context"
	"errors"
	"net/http"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var err error
var pod *corev1.Pod
var obj *v1alpha1.Quarantine

const quarantineControllerLabelKey = "component"
const quarantineControllerLabelValue = "incident-controller-manager"

// Handle handles admission requests.
func (h *QuarantineHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	if obj, err = h.manageObject(req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch t := req.Operation; t {
	case admissionv1.Create:
		err = h.ValidateCreate()
	case admissionv1.Update:
		err = h.ValidateUpdate(req.OldObject.Object)
	case admissionv1.Delete:
		err = h.ValidateDelete()
	}

	if err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("controller is on a valid node")
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineHandler) ValidateCreate() error {
	h.Log.Info("validate create", "name", obj.Name)

	if pod, err = h.getControllerPod(); err != nil {
		h.Log.Info("error on getting controller pod")
		return err
	}

	if ok := h.controllerShouldBeRescheduled(pod.Spec.NodeName); ok {
		h.Log.Info("controller pod is on a node marked for isolation")
		return errors.New("controller pod is on a node marked for isolation")
	}

	h.Log.Info("controller pod is on a valid node")

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineHandler) ValidateUpdate(old runtime.Object) error {
	h.Log.Info("validate update", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineHandler) ValidateDelete() error {
	h.Log.Info("validate delete", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (h *QuarantineHandler) controllerShouldBeRescheduled(nodeName string) bool {

	for _, n := range obj.Spec.Nodes {
		if n.Name == nodeName {
			return true
		}
	}

	return false
}

func (h *QuarantineHandler) manageObject(req admission.Request) (*v1alpha1.Quarantine, error) {

	quarantine := &v1alpha1.Quarantine{}

	if err := h.Decoder.Decode(req, quarantine); err != nil {
		return quarantine, err
	}

	return quarantine, nil
}

func (h *QuarantineHandler) getControllerPod() (*corev1.Pod, error) {

	pods := &corev1.PodList{}
	var pod *corev1.Pod
	var err error

	selector, err := labels.Parse(quarantineControllerLabelKey + "=" + quarantineControllerLabelValue)

	if err != nil {
		return pod, err
	}

	listOpts := client.ListOptions{
		LabelSelector: selector,
	}

	if err = h.Client.List(context.TODO(), pods, &listOpts); err != nil {
		return pod, err
	}

	if len(pods.Items) == 0 {
		return pod, errors.New("no controller pod found")
	}

	if len(pods.Items) > 1 {
		return pod, errors.New("multiple controller pods found")
	}

	return &pods.Items[0], nil
}
