package quarantine

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

const QuarantinePodLabel = "quarantine=true"

func updatePod(c kubernetes.Interface, matchedLabels map[string]string, nodeName, namespace string) error {

	var pods *corev1.PodList
	var err error

	// define selector for getting wanted pod
	selectorStringList := []string{}

	for k, v := range matchedLabels {
		selectorStringList = append(selectorStringList, k+"="+v)
	}

	selectorStringList = append(selectorStringList, QuarantinePodLabel)

	listOpts := metav1.ListOptions{
		LabelSelector: strings.Join(selectorStringList, ","),
	}

	if pods, err = c.CoreV1().Pods(namespace).List(context.TODO(), listOpts); err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == nodeName {

			labels := map[string]string{}
			updateOpts := metav1.UpdateOptions{}

			for k := range pod.ObjectMeta.Labels {
				labels[k] = QuarantinePodSelector
			}

			pod.ObjectMeta.Labels = labels
			pod.Spec.Tolerations = append(pod.Spec.Tolerations, corev1.Toleration{
				Key:    QuarantineTaintKey,
				Value:  QuarantineTaintValue,
				Effect: QuarantineTaintEffect,
			})

			if _, err := c.CoreV1().Pods(namespace).Update(context.TODO(), &pod, updateOpts); err != nil {
				return err
			}
		}
	}

	return nil
}

func cleanupIsolatedPods(c kubernetes.Interface) error {

	var pods *corev1.PodList
	var err error

	listOpts := metav1.ListOptions{
		LabelSelector: QuarantinePodLabel,
	}

	if pods, err = c.CoreV1().Pods("").List(context.TODO(), listOpts); err != nil {
		return err
	}

	deleteOpts := metav1.DeleteOptions{}

	for _, pod := range pods.Items {
		if err = c.CoreV1().Pods(pod.ObjectMeta.Namespace).Delete(context.TODO(), pod.ObjectMeta.Name, deleteOpts); err != nil {
			return err
		}
	}

	return nil
}
