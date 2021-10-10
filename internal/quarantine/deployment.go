package quarantine

import (
	"context"
	"strings"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

func (d Deployment) isolatePod(c kubernetes.Interface, node string, isolatedNode bool) error {

	var obj *v1.Deployment
	var err error

	opts := metav1.GetOptions{}

	if obj, err = c.AppsV1().Deployments(d.Namespace).Get(context.Background(), d.Name, opts); err != nil {
		return err
	}

	if err := updatePod(c, obj.Spec.Selector.MatchLabels, node, d.Namespace, true, true); err != nil {
		return err
	}

	if isolatedNode {
		if err = updatePod(c, obj.Spec.Selector.MatchLabels, node, d.Namespace, false, true); err != nil {
			return err
		}
	}

	return nil
}

func (d Deployment) isAlreadyManaged(c kubernetes.Interface, node, namespace string) (bool, error) {

	var obj *v1.Deployment
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if obj, err = c.AppsV1().Deployments(d.Namespace).Get(context.TODO(), d.Name, getOpts); err != nil {
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
