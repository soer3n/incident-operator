package quarantine

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func (ds Daemonset) isolatePod(c kubernetes.Interface, isolatedNode bool) error {

	var obj *v1.DaemonSet
	var pods *corev1.PodList
	var err error

	getOpts := metav1.GetOptions{}
	patchOpts := metav1.PatchOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().DaemonSets(ds.Namespace).Get(context.Background(), ds.Name, getOpts); err != nil {
		return err
	}

	// define selector for getting wanted pod
	listOpts := metav1.ListOptions{
		LabelSelector: obj.Spec.Selector.String(),
	}

	if pods, err = c.CoreV1().Pods(ds.Namespace).List(context.Background(), listOpts); err != nil {
		return err
	}

	if isolatedNode {
		renderedLabels := ""
		patch := []byte(`{"spec":{"template":{"metadata": {"labels": "` + renderedLabels + "" + `"}}}}`)

		if _, err = c.CoreV1().Pods(ds.Namespace).Patch(context.Background(),
			pods.Items[0].ObjectMeta.Name,
			types.MergePatchType,
			patch, patchOpts); err != nil {
			return err
		}

		return nil
	}

	patch := []byte(`{"spec":{"template":{"spec": {"tolerations": [{"key": "` + QuarantineTaintKey + `", "operator": "Equal", "value": "` + QuarantineTaintValue + `", "effect": "NoSchedule"}]}}}}`)

	if _, err = c.AppsV1().DaemonSets(ds.Namespace).Patch(context.Background(),
		pods.Items[0].ObjectMeta.Name,
		types.MergePatchType,
		patch, patchOpts); err != nil {
		return err
	}

	return nil
}
