package cli

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"

	"github.com/soer3n/yaho/pkg/client"
	"sigs.k8s.io/descheduler/pkg/descheduler/evictions"
)

const (
	rescheduleStrategy  = "evict"
	evictionKind        = "Eviction"
	evictionSubresource = "pods/eviction"
)

// RescheduleQuarantineController represents descheduling of quarantine controller if needed due to validation
func RescheduleQuarantineController(excludedNodes []string) error {

	var err error
	var success bool
	var policyGroupVersion string
	var excludedNodesObj []*corev1.Node
	var pod *corev1.Pod
	var node *corev1.Node

	utilsClient := client.New()
	typedClient := utilsClient.TypedClient
	discoveryClient := utilsClient.DiscoverClient

	if pod, err = GetControllerPod(typedClient); err != nil {
		return err
	}

	if node, err = GetControllerNode(typedClient, pod); err != nil {
		return err
	}

	if err = labelNodes(typedClient, excludedNodes); err != nil {
		return err
	}

	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}

	pod.Spec.NodeSelector[quarantineControllerLabelKey] = quarantineControllerLabelValue

	if policyGroupVersion, err = supportEviction(discoveryClient); err != nil {
		return err
	}

	ev := evictions.NewPodEvictor(typedClient, policyGroupVersion, false, 1, excludedNodesObj, false, false, true)

	if success, err = ev.EvictPod(context.TODO(), pod, node, rescheduleStrategy); err != nil {
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
			if node.ObjectMeta.Name != e {
				label = false
			}

			if label {
				node.ObjectMeta.Labels[quarantineControllerLabelKey] = quarantineControllerLabelValue
				node.ObjectMeta.Labels[quarantineControllerLabelIgnoreNodeKey] = quarantineControllerLabelIgnoreNodeValue
				if _, err = c.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func supportEviction(client discovery.ServerResourcesInterface) (string, error) {
	groupList, _, err := client.ServerGroupsAndResources()
	if err != nil {
		return "", err
	}
	foundPolicyGroup := false
	var policyGroupVersion string
	for _, group := range groupList {
		if group.Name == "policy" {
			foundPolicyGroup = true
			policyGroupVersion = group.PreferredVersion.GroupVersion
			break
		}
	}
	if !foundPolicyGroup {
		return "", nil
	}
	resourceList, err := client.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return "", err
	}
	for _, resource := range resourceList.APIResources {
		if resource.Name == evictionSubresource && resource.Kind == evictionKind {
			return policyGroupVersion, nil
		}
	}
	return "", nil
}
