package cli

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/soer3n/yaho/pkg/client"
	"sigs.k8s.io/descheduler/pkg/descheduler/evictions"
)

const rescheduleStrategy = ""

// RescheduleQuarantineController represents descheduling of quarantine controller if needed due to validation
func RescheduleQuarantineController(excludedNodes []string) error {

	var err error
	var success bool
	var excludedNodesObj []*corev1.Node
	var pod *corev1.Pod
	var node *corev1.Node

	typedClient := client.New().TypedClient

	if pod, err = GetControllerPod(typedClient); err != nil {
		return err
	}

	if node, err = GetControllerNode(typedClient, pod); err != nil {
		return err
	}

	if err = labelNodes(typedClient, excludedNodes); err != nil {
		return err
	}

	pod.Spec.NodeSelector[quarantineControllerLabelKey] = quarantineControllerLabelValue

	ev := evictions.NewPodEvictor(typedClient, rescheduleStrategy, false, 1, excludedNodesObj, false)

	if success, err = ev.EvictPod(context.TODO(), pod, node); err != nil {
		return err
	}

	if !success {
		return errors.New("no success on rescheduling quarantine controller")
	}

	return nil
}

func labelNodes(c kubernetes.Interface, excludedNodes []string) error {

	var nodes *corev1.NodeList
	var node corev1.Node
	var err error

	listOpts := metav1.ListOptions{}

	if nodes, err = c.CoreV1().Nodes().List(context.TODO(), listOpts); err != nil {
		return err
	}

	for _, node = range nodes.Items {
		label := true
		for _, e := range excludedNodes {
			if node.ObjectMeta.Name == e {
				label = false
			}

			if label {
				node.ObjectMeta.Labels[quarantineControllerLabelKey] = quarantineControllerLabelValue
				if _, err = c.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
