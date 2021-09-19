package webhook

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

func (qh *QuarantineHandler) parseAdmissionResponse() error {

	if err := json.Unmarshal(qh.body, qh.response); err != nil {
		return err
	}

	return nil
}

func (qh *QuarantineHandler) controllerShouldBeRescheduled(pod *corev1.Pod) bool {
	return true
}

func (qh *QuarantineHandler) rescheduleController() error {
	return nil
}
