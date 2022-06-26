package quarantine

import (
	"context"
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/descheduler/pkg/descheduler/evictions"

	"k8s.io/client-go/kubernetes"
)

const QuarantinePodLabelPrefix = "ops.soer3n.info/"
const QuarantinePodLabelKey = "quarantine"
const QuarantineNodeRemoveLabel = "revert"
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

			currentPod := &corev1.Pod{}
			pod.DeepCopyInto(currentPod)

			updateOpts := metav1.UpdateOptions{}

			if updateLabels {
				labels := map[string]string{}
				labels[QuarantinePodLabelPrefix+QuarantinePodLabelKey] = quarantinePodLabelValue
				currentPod.ObjectMeta.Labels = labels
			}

			if addToleration {
				currentPod.Spec.Tolerations = append(pod.Spec.Tolerations, corev1.Toleration{
					Key:      quarantineTaintKey,
					Operator: quarantineTaintOperator,
					Effect:   quarantineTaintEffect,
				})
			}

			if _, err = c.CoreV1().Pods(namespace).Update(context.TODO(), currentPod, updateOpts); err != nil {
				return err
			}
		}
	}

	return nil
}

func podIsNotInQuarantine(pod corev1.Pod) bool {
	if _, ok := pod.ObjectMeta.Labels[QuarantinePodLabelPrefix+QuarantinePodLabelKey]; !ok {
		return true
	}

	return false
}

func cleanupIsolatedPods(c kubernetes.Interface) error {

	var pods *corev1.PodList
	var err error

	listOpts := metav1.ListOptions{
		LabelSelector: QuarantinePodLabelPrefix + QuarantinePodLabelKey + "=" + quarantinePodLabelValue,
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

func evictPod(pod corev1.Pod, c kubernetes.Interface) error {

	var err error
	var success bool
	var excludedNodesObj []*corev1.Node
	var node *corev1.Node

	ev := evictions.NewPodEvictor(c, "", false, 1, excludedNodesObj, false, false, true)

	if success, err = ev.EvictPod(context.TODO(), &pod, node, rescheduleStrategy); err != nil {
		return err
	}

	if !success {
		return errors.New("no success on pod eviction")
	}

	return nil
}
