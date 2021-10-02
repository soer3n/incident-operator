package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (qh *QuarantineHandler) parseAdmissionResponse() error {

	if err := json.Unmarshal(qh.body, qh.response); err != nil {
		return err
	}

	return nil
}

func (qh *QuarantineHandler) controllerShouldBeRescheduled(pod *corev1.Pod, q *v1alpha1.Quarantine) bool {

	for _, n := range q.Spec.Nodes {
		if n.Name == pod.Spec.NodeName {
			qh.response.Response.Allowed = false
			qh.response.Response.Result = &metav1.Status{
				Message: "Quarantine Controller is currently running on a node which is requested for isolation",
			}
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

	qh.response = arRequest

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
