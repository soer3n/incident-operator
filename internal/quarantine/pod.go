package quarantine

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

const quarantinePodLabelPrefix = "ops.soer3n.info/"
const quarantinePodLabelKey = "quarantine"
const quarantinePodLabelValue = "true"

func updatePod(c kubernetes.Interface, matchedLabels map[string]string, nodeName, namespace string, updateLabels, addToleration bool) error {

	var pods *corev1.PodList
	var err error

	// define selector for getting wanted pod
	selectorStringList := []string{}

	for k, v := range matchedLabels {
		selectorStringList = append(selectorStringList, k+"="+v)
	}

	listOpts := metav1.ListOptions{
		LabelSelector: strings.Join(selectorStringList, ","),
	}

	if pods, err = c.CoreV1().Pods(namespace).List(context.TODO(), listOpts); err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == nodeName {

			var currentPod *corev1.Pod
			pod.DeepCopyInto(currentPod)

			updateOpts := metav1.UpdateOptions{}

			if updateLabels {
				labels := map[string]string{}

				for k, v := range currentPod.ObjectMeta.Labels {
					labels[k] = quarantinePodSelector
					labels[quarantinePodLabelPrefix+k] = v
				}

				labels[quarantinePodLabelPrefix+quarantinePodLabelKey] = quarantinePodLabelValue

				currentPod.ObjectMeta.Labels = labels
			}

			if addToleration {
				currentPod.Spec.Tolerations = append(pod.Spec.Tolerations, corev1.Toleration{
					Key:    quarantineTaintKey,
					Value:  quarantineTaintValue,
					Effect: quarantineTaintEffect,
				})
			}

			if _, err = c.CoreV1().Pods(namespace).Update(context.TODO(), currentPod, updateOpts); err != nil {
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
		LabelSelector: quarantinePodLabelPrefix + quarantinePodLabelKey + "=" + quarantinePodLabelValue,
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
