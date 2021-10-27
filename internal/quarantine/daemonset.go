package quarantine

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes"
)

var obj *v1.DaemonSet
var patch []byte
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

		patchPayload := v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Tolerations: obj.Spec.Template.Spec.Tolerations,
				},
			},
		}

		patchOpts := metav1.PatchOptions{}

		if patch, err = json.Marshal(patchPayload); err != nil {
			return err
		}

		if _, err = c.AppsV1().DaemonSets(ds.Namespace).Patch(context.TODO(), ds.Name, types.StrategicMergePatchType, patch, patchOpts); err != nil {
			return err
		}

		labels, _ := ds.getLabelSelectorAsString(podMatchLabels)
		timeout := int64(20)
		listOpts := metav1.ListOptions{
			Watch:          true,
			LabelSelector:  labels,
			TimeoutSeconds: &timeout,
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

	patchPayload := v1.DaemonSetSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Tolerations: tolerations,
			},
		},
	}

	patchOpts := metav1.PatchOptions{}

	if patch, err = json.Marshal(patchPayload); err != nil {
		return err
	}

	log.Println(string(patch))

	if _, err = c.AppsV1().DaemonSets(ds.Namespace).Patch(context.TODO(), ds.Name, types.StrategicMergePatchType, patch, patchOpts); err != nil {
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
