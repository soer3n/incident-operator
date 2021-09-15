package quarantine

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func (ds Daemonset) isolatePod(c kubernetes.Interface, node string, isolatedNode bool) error {

	var obj *v1.DaemonSet
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, getOpts); err != nil {
		return err
	}

	if err = updatePod(c, obj.Spec.Selector.MatchLabels, node, ds.Namespace); err != nil {
		return err
	}

	if isolatedNode {
		patch := []byte(`{"spec":{"template":{"spec": {"tolerations": [{"key": "` + quarantineTaintKey + `", "operator": "Equal", "value": "` + quarantineTaintValue + `", "effect": "` + quarantineTaintEffect + `"}]}}}}`)

		if _, err = c.AppsV1().DaemonSets(ds.Namespace).Patch(context.Background(),
			ds.Name,
			types.MergePatchType,
			patch, metav1.PatchOptions{}); err != nil {
			return err
		}

		return nil
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
