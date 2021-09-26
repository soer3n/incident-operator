package quarantine

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/kubectl/pkg/drain"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/yaho/pkg/client"
)

const dsType = "daemonset"
const deploymentType = "deployment"

func (n Node) prepare() error {

	for _, ds := range n.Daemonsets {

		if err := ds.isolatePod(client.New().TypedClient, n.Name, n.isolate); err != nil {
			return err
		}
	}

	for _, d := range n.Deployments {

		if err := d.isolatePod(client.New().TypedClient, n.Name); err != nil {
			return err
		}
	}

	if n.isolate {
		if err := n.addTaint(); err != nil {
			return err
		}
	}

	if err := n.disableScheduling(); err != nil {
		return err
	}

	return nil
}

func (n *Node) update() error {

	for _, ds := range n.Daemonsets {

		ok, err := ds.isAlreadyManaged(n.flags.Client, n.Name, ds.Namespace)

		if err != nil {
			return err
		}

		if !ok {
			if err := ds.isolatePod(n.flags.Client, n.Name, n.isolate); err != nil {
				return err
			}
		}
	}

	for _, d := range n.Deployments {

		ok, err := d.isAlreadyManaged(n.flags.Client, n.Name, d.Namespace)

		if err != nil {
			return err
		}

		if !ok {
			if err := d.isolatePod(n.flags.Client, n.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) mergeResources(rs []v1alpha1.Resource) error {

	for _, r := range rs {
		switch t := r.Type; t {
		case dsType:
			for _, v := range n.Daemonsets {
				if v.Name == r.Name && v.Namespace == r.Namespace {
					continue
				}
				n.Daemonsets = append(n.Daemonsets, Daemonset{
					Name:      v.Name,
					Namespace: v.Namespace,
				})

			}

			n.Daemonsets = append(n.Daemonsets, Daemonset{
				Name:      r.Name,
				Namespace: r.Namespace,
			})
		case deploymentType:
			for _, v := range n.Deployments {
				if v.Name == r.Name && v.Namespace == r.Namespace {
					continue
				}
				n.Deployments = append(n.Deployments, Deployment{
					Name:      v.Name,
					Namespace: v.Namespace,
				})
			}

			n.Deployments = append(n.Deployments, Deployment{
				Name:      r.Name,
				Namespace: r.Namespace,
			})
		}
	}

	return nil
}

func (n *Node) parseFlags() {
	n.flags = &drain.Helper{
		IgnoreAllDaemonSets: true,
		DisableEviction:     false,
		PodSelector:         "!" + quarantinePodSelector,
		Force:               false,
		Ctx:                 context.TODO(),
		Client:              client.New().TypedClient,
		ErrOut:              n.ioStreams.ErrOut,
		Out:                 n.ioStreams.Out,
	}
}

func (n Node) disableScheduling() error {

	nodeObj := n.getNodeAPIObject()

	if err := drain.RunCordonOrUncordon(n.flags, nodeObj, true); err != nil {
		return err
	}

	return nil
}

func (n Node) enableScheduling() error {

	nodeObj := n.getNodeAPIObject()

	if err := drain.RunCordonOrUncordon(n.flags, nodeObj, false); err != nil {
		return err
	}

	return nil
}

func (n Node) addTaint() error {

	nodeObj := n.getNodeAPIObject()

	for _, taint := range nodeObj.Spec.Taints {
		if taint.Key == quarantineTaintKey && taint.Value == quarantineTaintValue {
			return nil
		}
	}

	nodeObj.Spec.Taints = append(nodeObj.Spec.Taints, corev1.Taint{
		Key:    quarantineTaintKey,
		Value:  quarantineTaintValue,
		Effect: quarantineTaintEffect,
	})

	if err := n.updateNodeAPIObject(nodeObj); err != nil {
		return err
	}

	return nil
}

func (n Node) removeTaint() error {

	nodeObj := n.getNodeAPIObject()
	taints := []corev1.Taint{}

	for _, taint := range nodeObj.Spec.Taints {
		if taint.Key != quarantineTaintKey && taint.Value != quarantineTaintValue {
			taints = append(taints, taint)
		}
	}

	nodeObj.Spec.Taints = taints

	if err := n.updateNodeAPIObject(nodeObj); err != nil {
		return err
	}

	return nil
}

func (n Node) deschedulePods() error {
	if err := drain.RunNodeDrain(n.flags, n.Name); err != nil {
		return err
	}

	return nil
}

func (n Node) getNodeAPIObject() *corev1.Node {

	var err error
	var nodeObj *corev1.Node

	opts := metav1.GetOptions{}

	if nodeObj, err = client.New().TypedClient.CoreV1().Nodes().Get(context.Background(), n.Name, opts); err != nil {
		println(err.Error())
	}

	return nodeObj
}

func (n Node) updateNodeAPIObject(nodeObj *corev1.Node) error {

	var err error

	opts := metav1.UpdateOptions{}

	if _, err = client.New().TypedClient.CoreV1().Nodes().Update(context.Background(), nodeObj, opts); err != nil {
		return err
	}

	return nil
}

func (n Node) isAlreadyIsolated() (bool, error) {

	nodeObj := n.getNodeAPIObject()
	return nodeObj.Spec.Unschedulable, nil
}
