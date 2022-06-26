package quarantine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/quarantine"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var err error
var pod *corev1.Pod

// var obj *v1alpha1.Quarantine

const quarantineControllerLabelKey = "control-plane"
const quarantineControllerLabelValue = "controller-manager"

// Handle handles admission requests.
func (h *QuarantineValidateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	var obj *v1alpha1.Quarantine

	if obj, err = h.manageObject(req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch t := req.Operation; t {
	case admissionv1.Create:
		err = h.Validate(obj)
	case admissionv1.Update:
		err = h.Validate(obj)
	}

	if err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("controller is on a valid node")
}

// Handle handles admission requests.
func (h *QuarantineMutateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	var rawObj, rawPatchedObj []byte

	var obj, oldObj, patchedObj *v1alpha1.Quarantine

	if obj, oldObj, err = h.manageObject(req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch t := req.Operation; t {
	case admissionv1.Update:
		patchedObj, err = h.MutateUpdate(obj, oldObj)
	}

	if err != nil {
		return admission.Denied(err.Error())
	}

	if rawObj, err = json.Marshal(obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if rawPatchedObj, err = json.Marshal(patchedObj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(rawObj, rawPatchedObj)
}

// Validate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineValidateHandler) Validate(obj *v1alpha1.Quarantine) error {
	h.Log.Info("validate", "name", obj.Name)

	if pod, err = h.getControllerPod(); err != nil {
		h.Log.Info("error on getting controller pod")
		return err
	}

	if ok := h.controllerShouldBeRescheduled(pod.Spec.NodeName, obj.Spec.Nodes); ok {
		h.Log.Info("controller pod is on a node marked for isolation")
		return errors.New("controller pod is on a node marked for isolation")
	}

	h.Log.Info("controller pod is on a valid node")

	return nil
}

// MutateUpdate implements webhook.Validator so a webhook will be registered for the type
func (h *QuarantineMutateHandler) MutateUpdate(obj, old *v1alpha1.Quarantine) (*v1alpha1.Quarantine, error) {
	h.Log.Info("mutate update", "name", obj.Name)

	patched := obj.DeepCopy()

	markedNodes := []string{}

	for _, on := range old.Spec.Nodes {

		isMarked := true
		for _, cn := range obj.Spec.Nodes {
			if on.Name == cn.Name {
				isMarked = false
				break
			}
		}

		if isMarked {
			markedNodes = append(markedNodes, on.Name)
		}
	}

	if len(markedNodes) == 0 {
		h.Log.Info("no nodes to remove from quarantine")
		return patched, nil
	}

	h.Log.Info("marked for removal from quarantine:", "nodes", markedNodes)

	if obj.ObjectMeta.Annotations == nil {
		patched.ObjectMeta.Annotations = map[string]string{}
	}

	patched.ObjectMeta.Annotations[quarantine.QuarantinePodLabelPrefix+quarantine.QuarantineNodeRemoveLabel] = strings.Join(markedNodes, ",")

	return patched, nil
}

func (h *QuarantineValidateHandler) controllerShouldBeRescheduled(nodeName string, nodes []v1alpha1.Node) bool {

	for _, n := range nodes {
		if n.Name == nodeName {
			return true
		}
	}

	return false
}

func (h *QuarantineValidateHandler) manageObject(req admission.Request) (*v1alpha1.Quarantine, error) {

	quarantine := &v1alpha1.Quarantine{}

	if err := h.Decoder.DecodeRaw(req.Object, quarantine); err != nil {
		return quarantine, err
	}

	return quarantine, nil
}

func (h *QuarantineMutateHandler) manageObject(req admission.Request) (*v1alpha1.Quarantine, *v1alpha1.Quarantine, error) {

	quarantine := &v1alpha1.Quarantine{}
	oldQuarantine := &v1alpha1.Quarantine{}

	if err := h.Decoder.DecodeRaw(req.Object, quarantine); err != nil {
		return quarantine, oldQuarantine, err
	}

	if err := h.Decoder.DecodeRaw(req.OldObject, oldQuarantine); err != nil {
		return quarantine, oldQuarantine, err
	}

	return quarantine, oldQuarantine, nil
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
