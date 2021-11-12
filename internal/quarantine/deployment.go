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

func (d Deployment) manageWorkload(c kubernetes.Interface, node string, isolatedNode bool, logger logr.Logger) error {

	if d.Keep {

		ok, err := d.isAlreadyManaged(c, node, d.Namespace)

		if err != nil {
			return err
		}

		if !ok {
			return errors.New("something went wrong on deployment " + d.Name)
		}
	}

	if ok, err := d.isAlreadyIsolated(c, node, d.Namespace); !ok {

		if err != nil {
			return err
		}

		if err := d.isolatePod(c, node, isolatedNode, logger); err != nil {
			return err
		}
	}

	return nil
}

func (d Deployment) isolatePod(c kubernetes.Interface, node string, isolatedNode bool, logger logr.Logger) error {

	var obj *v1.Deployment
	var patch []byte
	var err error

	opts := metav1.GetOptions{}

	if obj, err = c.AppsV1().Deployments(d.Namespace).Get(context.Background(), d.Name, opts); err != nil {
		return err
	}

	if err := updatePod(c, obj.Spec.Selector.MatchLabels, node, d.Namespace, true, true); err != nil {
		return err
	}

	logger.Info("pod isolated from workload...")

	if d.Keep {
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

		if _, err = c.AppsV1().Deployments(d.Namespace).Patch(context.TODO(), d.Name, types.JSONPatchType, patch, patchOpts); err != nil {
			return err
		}

		logger.Info("modified...")

	}

	return nil
}

func (d Deployment) removeToleration(c kubernetes.Interface) error {

	var patch []byte
	var err error

	getOpts := metav1.GetOptions{}

	// get affected daemonset
	if _, err = c.AppsV1().Deployments(d.Namespace).Get(context.TODO(), d.Name, getOpts); err != nil {
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

	if _, err = c.AppsV1().Deployments(d.Namespace).Patch(context.TODO(), d.Name, types.JSONPatchType, patch, patchOpts); err != nil {
		return err
	}

	patchLabelPayload := []labelPayload{
		{
			Op:   "add",
			Path: "/metadata/labels",
			Value: map[string]string{
				quarantinePodLabelPrefix + quarantinePodLabelKey: quarantinePodLabelValue,
			},
		},
	}

	if patch, err = json.Marshal(patchLabelPayload); err != nil {
		return err
	}

	if _, err = c.AppsV1().Deployments(d.Namespace).Patch(context.TODO(), d.Name, types.JSONPatchType, patch, patchOpts); err != nil {
		return err
	}

	return nil
}

func (d Deployment) isAlreadyIsolated(c kubernetes.Interface, node, namespace string) (bool, error) {

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
		if pod.ObjectMeta.Namespace == namespace && strings.Contains(pod.ObjectMeta.Name, d.Name) {
			return true, nil
		}
	}

	return false, nil
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
