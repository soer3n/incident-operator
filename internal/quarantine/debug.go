package quarantine

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

const debugPodName = "quarantine-debug"
const debugPodNamespace = "kube-system"
const debugPodImage = "nicolaka/netshoot"
const debugPodContainerName = "debug"

func (dg Debug) deploy(c kubernetes.Interface, nodeName string) error {
	var err error

	autoMountToken := new(bool)
	*autoMountToken = false
	debugPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: debugPodName,
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: autoMountToken,
			PriorityClassName:            "system-node-critical",
			HostNetwork:                  true,
			NodeName:                     nodeName,
			Tolerations: []corev1.Toleration{
				{
					Key:    QuarantineTaintKey,
					Value:  QuarantineTaintValue,
					Effect: QuarantineTaintEffect,
				},
			},
			Containers: []corev1.Container{
				{
					Name:  debugPodContainerName,
					Image: debugPodImage,
					Stdin: true,
					TTY:   true,
				},
			},
		},
	}
	createOpts := metav1.CreateOptions{}

	if _, err = c.CoreV1().Pods(dg.Namespace).Create(context.TODO(), debugPod, createOpts); err != nil {
		return err
	}

	return nil
}

func (dg Debug) remove(c kubernetes.Interface, name, namespace string) error {

	deleteOpts := metav1.DeleteOptions{}

	if err := c.CoreV1().Pods(namespace).Delete(context.TODO(), name, deleteOpts); err != nil {
		return err
	}

	return nil
}
