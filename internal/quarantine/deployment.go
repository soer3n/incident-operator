package quarantine

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (d Deployment) isolatePod(c kubernetes.Interface, node string) error {
	opts := metav1.GetOptions{}
	var obj *v1.Deployment
	var err error

	if obj, err = c.AppsV1().Deployments(d.Namespace).Get(context.Background(), d.Name, opts); err != nil {
		return err
	}

	return updatePod(c, obj.Spec.Selector.MatchLabels, node, d.Namespace)
}
