package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const quarantineControllerLabel = "control-plane=incident-controller-manager"

func (qh *QuarantineHandler) parseAdmissionResponse() error {

	if err := json.Unmarshal(qh.body, qh.response); err != nil {
		return err
	}

	return nil
}

func (qh QuarantineHandler) getControllerPod() (*corev1.Pod, error) {

	var pods *corev1.PodList
	var pod *corev1.Pod
	var err error

	listOpts := metav1.ListOptions{
		LabelSelector: quarantineControllerLabel,
	}

	if pods, err = qh.client.CoreV1().Pods("").List(context.TODO(), listOpts); err != nil {
		return pod, err
	}

	if len(pods.Items) > 1 {
		return pod, errors.New("multiple controller pods found")
	}

	return &pods.Items[0], nil
}

func (qh *QuarantineHandler) controllerShouldBeRescheduled(pod *corev1.Pod, q *v1alpha1.Quarantine) bool {

	for _, n := range q.Spec.Nodes {
		if n.Name == pod.Spec.NodeName {
			return true
		}
	}

	return false
}

func (qh *QuarantineHandler) getAdmissionRequestSpec(body []byte, w http.ResponseWriter) (*v1beta1.AdmissionReview, error) {

	arRequest := &v1beta1.AdmissionReview{}

	if err := json.Unmarshal(body, &arRequest); err != nil {
		return arRequest, err
	}

	return arRequest, nil
}

func (qh *QuarantineHandler) getQuarantineSpec(arRequest *v1beta1.AdmissionReview, w http.ResponseWriter) (*v1alpha1.Quarantine, error) {

	raw := arRequest.Request.Object.Raw
	quarantine := &v1alpha1.Quarantine{}

	if err := json.Unmarshal(raw, quarantine); err != nil {
		return quarantine, err
	}

	return quarantine, nil
}
