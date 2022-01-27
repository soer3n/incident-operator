package quarantine

import (
	"context"
	"encoding/json"
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
func (h *QuarantineValidateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	if obj, err = h.manageObject(req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch t := req.Operation; t {
	case admissionv1.Create:
		err = h.Validate()
	case admissionv1.Update:
		err = h.Validate()
	}

	if err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("controller is on a valid node")
}

// Handle handles admission requests.
func (h *QuarantineMutateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	oldObj := &v1alpha1.Quarantine{}

	if oldObj, err = h.manageObject(req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch t := req.Operation; t {
	case admissionv1.Update:
		err = h.MutateUpdate(oldObj)
	}

	if err != nil {
		return admission.Denied(err.Error())
	}

	rawObj, _ := json.Marshal(obj)
	return admission.PatchResponseFromRaw(rawObj, rawObj)
}

// Validate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineValidateHandler) Validate() error {
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

// MutateUpdate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineMutateHandler) MutateUpdate(old runtime.Object) error {
	h.Log.Info("validate update", "name", obj.Name)

	markedNodes := []*v1alpha1.Node{}

	oo := old.(*v1alpha1.Quarantine)

	for _, on := range oo.Spec.Nodes {

		isMarked := true
		for _, cn := range obj.Spec.Nodes {
			if on.Name == cn.Name {
				isMarked = false
				break
			}
		}

		if isMarked {
			markedNodes = append(markedNodes, &on)
		}
	}

	h.Log.Info("marked:", "nodes", markedNodes)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

func (h *QuarantineValidateHandler) controllerShouldBeRescheduled(nodeName string) bool {

	for _, n := range obj.Spec.Nodes {
		if n.Name == nodeName {
			return true
		}
	}

	return false
}

func (h *QuarantineValidateHandler) manageObject(req admission.Request) (*v1alpha1.Quarantine, error) {

	quarantine := &v1alpha1.Quarantine{}

	if err := h.Decoder.Decode(req, quarantine); err != nil {
		return quarantine, err
	}

	return quarantine, nil
}

func (h *QuarantineMutateHandler) manageObject(req admission.Request) (*v1alpha1.Quarantine, error) {

	quarantine := &v1alpha1.Quarantine{}

	if err := h.Decoder.Decode(req, quarantine); err != nil {
		return quarantine, err
	}

	return quarantine, nil
}

func (h *QuarantineValidateHandler) getControllerPod() (*corev1.Pod, error) {

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
