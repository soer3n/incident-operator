package quarantine

import (
	"context"
	"strings"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

var obj *v1.DaemonSet
var err error

func (ds Daemonset) isolatePod(c kubernetes.Interface, node string, isolatedNode bool, logger logr.Logger) error {

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return err
	}

	podMatchLabels := obj.Spec.Selector.DeepCopy()

	if err = updatePod(c, podMatchLabels.MatchLabels, node, ds.Namespace, true, true); err != nil {
		return err
	}

	if ds.Keep {
		obj.Spec.Template.Spec.Tolerations = append(obj.Spec.Template.Spec.Tolerations, corev1.Toleration{
			Key:    quarantineTaintKey,
			Value:  quarantineTaintValue,
			Effect: quarantineTaintEffect,
		})

		updateOpts := metav1.UpdateOptions{}

		if _, err = c.AppsV1().DaemonSets(ds.Namespace).Update(context.TODO(), obj, updateOpts); err != nil {
			return err
		}

		labels, _ := ds.getLabelSelectorAsString(podMatchLabels)
		listOpts := metav1.ListOptions{
			Watch:         true,
			LabelSelector: labels,
		}

		w, err := c.CoreV1().Pods(ds.Namespace).Watch(context.TODO(), listOpts)

		if err != nil {
			return err
		}

		waitForResource(w, logger)
	}

	return nil
}

func (ds Daemonset) removeToleration(c kubernetes.Interface) error {

	var obj *v1.DaemonSet
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return err
	}

	tolerations := []corev1.Toleration{}

	for _, t := range obj.Spec.Template.Spec.Tolerations {
		if t.Value != quarantineTaintValue && t.Key != quarantineTaintKey {
			tolerations = append(tolerations, t)
		}
	}

	obj.Spec.Template.Spec.Tolerations = tolerations
	updateOpts := metav1.UpdateOptions{}

	// get affected daemonset
	if _, err = c.AppsV1().DaemonSets(ds.Namespace).Update(context.TODO(), obj, updateOpts); err != nil {
		return err
	}

	return nil
}

func (ds Daemonset) isAlreadyManaged(c kubernetes.Interface, node, namespace string) (bool, error) {

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return false, err
	}

	labels, _ := ds.getLabelSelectorAsString(obj.Spec.Selector)
	labels = labels + ",kubernetes.io/hostname=" + node

	listOpts := metav1.ListOptions{
		LabelSelector: labels,
	}

	var podList *corev1.PodList

	if podList, err = c.CoreV1().Pods(namespace).List(context.TODO(), listOpts); err != nil {
		return false, err
	}

	if len(podList.Items) < 1 {
		return false, nil
	}

	return true, nil
}

func (ds Daemonset) getLabelSelectorAsString(podMatchLabels *metav1.LabelSelector) (string, error) {
	// define selector for getting wanted pod
	selectorStringList := []string{}

	for k, v := range podMatchLabels.MatchLabels {
		selectorStringList = append(selectorStringList, quarantinePodLabelPrefix+k+"="+v)
	}

	return strings.Join(selectorStringList, ","), nil
}
