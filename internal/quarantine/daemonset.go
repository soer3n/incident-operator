package quarantine

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes"
)

const (
	rescheduleStrategy = "evict"
)

func (ds Daemonset) manageWorkload(c kubernetes.Interface, node string, isolatedNode bool, logger logr.Logger) error {

	if ds.Keep {

		ok, err := ds.isAlreadyManaged(c, node, ds.Namespace)

		if err != nil {
			return err
		}

		if !ok {
			return errors.New("something went wrong on daemonset " + ds.Name)
		}
	}

	if ok, err := ds.isAlreadyIsolated(c, node, ds.Namespace); !ok {

		if err != nil {
			return err
		}

		if err := ds.isolatePod(c, node, isolatedNode, logger); err != nil {
			return err
		}
	}

	return nil
}

func (ds Daemonset) isolatePod(c kubernetes.Interface, node string, isolatedNode bool, logger logr.Logger) error {

	var obj *v1.DaemonSet
	var patch []byte
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return err
	}

	podMatchLabels := obj.Spec.Selector.DeepCopy()

	if err = updatePod(c, podMatchLabels.MatchLabels, node, ds.Namespace, true, true); err != nil {
		return err
	}

	logger.Info("pod isolated from workload...")

	if ds.Keep {

		patchPayload := []tolerationPayload{
			{
				Op:   "add",
				Path: "/spec/template/spec/tolerations",
				Value: []tolerationValue{
					{
						Key:      quarantineTaintKey,
						Operator: quarantineTaintOperator,
						Effect:   quarantineTaintEffect,
					},
				},
			},
		}

		patchOpts := metav1.PatchOptions{}

		if patch, err = json.Marshal(patchPayload); err != nil {
			return err
		}

		if _, err = c.AppsV1().DaemonSets(ds.Namespace).Patch(context.TODO(), ds.Name, types.JSONPatchType, patch, patchOpts); err != nil {
			return err
		}

		logger.Info("modified...")
	}

	return nil
}

func (ds Daemonset) removeToleration(c kubernetes.Interface) error {

	var patch []byte
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if _, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return err
	}

	patchPayload := []tolerationPayload{
		{
			Op:    "replace",
			Path:  "/spec/template/spec/tolerations",
			Value: []tolerationValue{},
		},
	}

	patchOpts := metav1.PatchOptions{}

	if patch, err = json.Marshal(patchPayload); err != nil {
		return err
	}

	if _, err = c.AppsV1().DaemonSets(ds.Namespace).Patch(context.TODO(), ds.Name, types.JSONPatchType, patch, patchOpts); err != nil {
		return err
	}

	return nil
}

func (ds Daemonset) isAlreadyIsolated(c kubernetes.Interface, node, namespace string) (bool, error) {

	var podList *corev1.PodList
	var err error

	core := c.CoreV1()

	listOpts := metav1.ListOptions{
		LabelSelector: quarantinePodLabelPrefix + quarantinePodLabelKey + "=" + quarantinePodLabelValue,
	}

	if podList, err = core.Pods(namespace).List(context.TODO(), listOpts); err != nil {
		return false, err
	}

	for _, pod := range podList.Items {
		if pod.ObjectMeta.Namespace == namespace && strings.Contains(pod.ObjectMeta.Name, ds.Name) {
			return true, nil
		}
	}

	return false, nil
}

func (ds Daemonset) isAlreadyManaged(c kubernetes.Interface, node, namespace string) (bool, error) {

	var obj *v1.DaemonSet
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return false, err
	}

	labels, _ := ds.getLabelSelectorAsString(obj.Spec.Selector)
	labels = labels + ""

	listOpts := metav1.ListOptions{
		LabelSelector: labels,
	}

	var podList *corev1.PodList
	core := c.CoreV1()
	if podList, err = core.Pods(namespace).List(context.TODO(), listOpts); err != nil {
		return false, err
	}

	if len(podList.Items) < 1 {
		return false, nil
	}

	for _, pods := range podList.Items {
		if pods.Spec.NodeName == node {
			return true, nil
		}
	}

	return false, nil
}

func (ds Daemonset) getLabelSelectorAsString(podMatchLabels *metav1.LabelSelector) (string, error) {
	// define selector for getting wanted pod
	selectorStringList := []string{}

	for k, v := range podMatchLabels.MatchLabels {
		selectorStringList = append(selectorStringList, quarantinePodLabelPrefix+k+"="+v)
	}

	return strings.Join(selectorStringList, ","), nil
}
