package utils

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const quarantineControllerLabel = "control-plane=incident-controller-manager"

// Contains represents func for checking if a string is in a list of strings
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func GetControllerPod(c kubernetes.Interface) (*corev1.Pod, error) {

	var pods *corev1.PodList
	var pod *corev1.Pod
	var err error

	listOpts := metav1.ListOptions{
		LabelSelector: quarantineControllerLabel,
	}

	if pods, err = c.CoreV1().Pods("").List(context.TODO(), listOpts); err != nil {
		return pod, err
	}

	if len(pods.Items) > 1 {
		return pod, errors.New("multiple controller pods found")
	}

	return &pods.Items[0], nil
}

func GetControllerNode(c kubernetes.Interface, pod *corev1.Pod) (*corev1.Node, error) {

	var node *corev1.Node
	var err error

	getOpts := metav1.GetOptions{}

	if node, err = c.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, getOpts); err != nil {
		return node, err
	}

	return node, nil
}
