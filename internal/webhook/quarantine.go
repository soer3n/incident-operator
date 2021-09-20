package webhook

import (
	"context"
	"encoding/json"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	// define selector for getting wanted pod
	selectorStringList := []string{}

	for k, v := range matchedLabels {
		selectorStringList = append(selectorStringList, k+"="+v)
	}

	listOpts := metav1.ListOptions{
		LabelSelector: strings.Join(selectorStringList, ","),
	}

	if pods, err = qh.client.CoreV1().Pods("").List(context.TODO(), listOpts); err != nil {
		return pod, err
	}

	return pod, nil
}

func (qh *QuarantineHandler) controllerShouldBeRescheduled(pod *corev1.Pod) bool {
	return true
}

func (qh *QuarantineHandler) rescheduleController() error {
	return nil
}
