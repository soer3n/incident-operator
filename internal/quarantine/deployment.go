package quarantine

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func (d Deployment) isolatePod(c kubernetes.Interface) error {
	opts := metav1.GetOptions{}
	var obj *v1.Deployment
	var pods *corev1.PodList
	var err error

	if obj, err = c.AppsV1().Deployments(d.Namespace).Get(context.Background(), d.Name, opts); err != nil {
		return err
	}

	// define selector for getting wanted pod
	listOpts := metav1.ListOptions{
		LabelSelector: obj.Spec.Selector.String(),
	}

	if pods, err = c.CoreV1().Pods(d.Namespace).List(context.Background(), listOpts); err != nil {
		return err
	}

	renderedLabels := ""
	patch := []byte(`{"spec":{"template":{"metadata": {"labels": "` + renderedLabels + `"}}}}`)
	patchOpts := metav1.PatchOptions{}

	if _, err = c.CoreV1().Pods(d.Namespace).Patch(context.Background(),
		pods.Items[0].ObjectMeta.Name,
		types.MergePatchType,
		patch, patchOpts); err != nil {
		return err
	}

	return nil
}
