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

	getOpts := metav1.GetOptions{}

	if _, err = c.CoreV1().Pods(dg.Namespace).Get(context.TODO(), debugPodName, getOpts); err == nil {
		return nil
	}

	autoMountToken := new(bool)
	*autoMountToken = false
	debugPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   debugPodName,
			Labels: map[string]string{},
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: autoMountToken,
			PriorityClassName:            "system-node-critical",
			HostNetwork:                  true,
			NodeName:                     nodeName,
			Tolerations: []corev1.Toleration{
				{
					Key:    quarantineTaintKey,
					Value:  quarantineTaintValue,
					Effect: quarantineTaintEffect,
				},
			},
			Containers: []corev1.Container{
				{
					Name:  debugPodContainerName,
					Image: debugPodImage,
					Stdin: true,
					TTY:   true,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host-system",
							ReadOnly:  true,
							MountPath: "/host",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "host-system",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/",
						},
					},
				},
			},
		},
	}

	debugPod.ObjectMeta.Labels[quarantinePodLabelPrefix+quarantinePodLabelKey] = quarantinePodLabelValue
	createOpts := metav1.CreateOptions{}

	if _, err = c.CoreV1().Pods(dg.Namespace).Create(context.TODO(), debugPod, createOpts); err != nil {
		return err
	}

	return nil
}

func (dg Debug) remove(c kubernetes.Interface, name, namespace string) error {

	var err error

	getOpts := metav1.GetOptions{}

	if _, err = c.CoreV1().Pods(namespace).Get(context.TODO(), name, getOpts); err != nil {
		return err
	}

	deleteOpts := metav1.DeleteOptions{}

	if err = c.CoreV1().Pods(namespace).Delete(context.TODO(), name, deleteOpts); err != nil {
		return err
	}

	return nil
}
