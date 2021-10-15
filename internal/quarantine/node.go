package quarantine

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/kubectl/pkg/drain"

	"github.com/soer3n/incident-operator/api/v1alpha1"
)

const dsType = "daemonset"
const deploymentType = "deployment"

func (n Node) prepare() error {

	for _, ds := range n.Daemonsets {

		ok, err := ds.isAlreadyManaged(n.Flags.Client, n.Name, ds.Namespace)

		if err != nil {
			return err
		}

		if ok {
			continue
		}

		if err := ds.isolatePod(n.Flags.Client, n.Name, n.isolate, n.logger.WithValues("daemonset", ds.Name)); err != nil {
			return err
		}
	}

	for _, d := range n.Deployments {

		if err := d.isolatePod(n.Flags.Client, n.Name, n.isolate, n.logger.WithValues("deployment", d.Name)); err != nil {
			return err
		}
	}

	if err := n.disableScheduling(); err != nil {
		return err
	}

	if n.isolate {
		if err := n.addTaint(); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) update() error {

	for _, ds := range n.Daemonsets {

		ok, err := ds.isAlreadyManaged(n.Flags.Client, n.Name, ds.Namespace)

		if err != nil {
			return err
		}

		if !ok {
			if err := ds.isolatePod(n.Flags.Client, n.Name, n.isolate, n.logger.WithValues("daemonset", ds.Name)); err != nil {
				return err
			}
		}
	}

	for _, d := range n.Deployments {

		ok, err := d.isAlreadyManaged(n.Flags.Client, n.Name, d.Namespace)

		if err != nil {
			return err
		}

		if !ok {
			if err := d.isolatePod(n.Flags.Client, n.Name, n.isolate, n.logger.WithValues("deployment", d.Name)); err != nil {
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
					Keep:      v.Keep,
				})

			}

			n.Daemonsets = append(n.Daemonsets, Daemonset{
				Name:      r.Name,
				Namespace: r.Namespace,
				Keep:      r.Keep,
			})
		case deploymentType:
			for _, v := range n.Deployments {
				if v.Name == r.Name && v.Namespace == r.Namespace {
					continue
				}
				n.Deployments = append(n.Deployments, Deployment{
					Name:      v.Name,
					Namespace: v.Namespace,
					Keep:      v.Keep,
				})
			}

			n.Deployments = append(n.Deployments, Deployment{
				Name:      r.Name,
				Namespace: r.Namespace,
				Keep:      r.Keep,
			})
		}
	}

	return nil
}

func (n *Node) parseFlags(c kubernetes.Interface) {

	if err != nil {
		n.logger.Error(err, "failure on init node drain helper")
	}

	n.Flags = &drain.Helper{
		IgnoreAllDaemonSets: true,
		DisableEviction:     false,
		DeleteEmptyDirData:  true,
		PodSelector:         "!" + quarantinePodLabelPrefix + quarantinePodSelector,
		Force:               false,
		Ctx:                 context.TODO(),
		Client:              c,
		ErrOut:              n.ioStreams.ErrOut,
		Out:                 n.ioStreams.Out,
	}
}

func (n Node) disableScheduling() error {

	nodeObj := n.getNodeAPIObject()

	if err := drain.RunCordonOrUncordon(n.Flags, nodeObj, true); err != nil {
		return err
	}

	return nil
}

func (n Node) enableScheduling() error {

	nodeObj := n.getNodeAPIObject()

	if err := drain.RunCordonOrUncordon(n.Flags, nodeObj, false); err != nil {
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
	if err := drain.RunNodeDrain(n.Flags, n.Name); err != nil {
		return err
	}

	return nil
}

func (n Node) getNodeAPIObject() *corev1.Node {

	var err error
	var nodeObj *corev1.Node

	opts := metav1.GetOptions{}

	if nodeObj, err = n.Flags.Client.CoreV1().Nodes().Get(context.Background(), n.Name, opts); err != nil {
		println(err.Error())
	}

	return nodeObj
}

func (n Node) updateNodeAPIObject(nodeObj *corev1.Node) error {

	var err error

	opts := metav1.UpdateOptions{}

	if _, err = n.Flags.Client.CoreV1().Nodes().Update(context.TODO(), nodeObj, opts); err != nil {
		return err
	}

	listOpts := metav1.ListOptions{
		Watch:         true,
		LabelSelector: "kubernetes.io/hostname=" + n.Name,
	}

	w, err := n.Flags.Client.CoreV1().Nodes().Watch(context.TODO(), listOpts)

	if err != nil {
		return err
	}

	waitForResource(w, n.logger)

	return nil
}

func (n Node) isAlreadyIsolated() (bool, error) {

	nodeObj := n.getNodeAPIObject()
	return nodeObj.Spec.Unschedulable, nil
}
