package quarantine

import (
	"context"
	"strings"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

var obj *v1.DaemonSet
var err error

func (ds Daemonset) isolatePod(c kubernetes.Interface, node string, isolatedNode bool) error {

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return err
	}

	if err = updatePod(c, obj.Spec.Selector.MatchLabels, node, ds.Namespace, true, true); err != nil {
		return err
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

	// define selector for getting wanted pod
	selectorStringList := []string{}

	for k, v := range obj.Spec.Selector.MatchLabels {
		selectorStringList = append(selectorStringList, quarantinePodLabelPrefix+k+"="+v)
	}

	selectorStringList = append(selectorStringList, "kubernetes.io/hostname="+node)

	listOpts := metav1.ListOptions{
		LabelSelector: strings.Join(selectorStringList, ","),
	}

	if _, err = c.CoreV1().Pods(namespace).List(context.TODO(), listOpts); err != nil {
		return false, err
	}

	return true, nil
}
