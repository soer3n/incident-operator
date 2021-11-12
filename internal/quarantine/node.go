package quarantine

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/kubectl/pkg/drain"

	"github.com/soer3n/incident-operator/api/v1alpha1"
)

const dsType = "daemonset"
const deploymentType = "deployment"

func (n Node) manageWorkloads() error {

	for _, ds := range n.Daemonsets {

		if err := ds.manageWorkload(n.Flags.Client, n.Name, n.Isolate, n.Logger.WithValues("daemonset", ds.Name)); err != nil {
			return err
		}
	}

	for _, d := range n.Deployments {

		if err := d.manageWorkload(n.Flags.Client, n.Name, n.Isolate, n.Logger.WithValues("deployment", d.Name)); err != nil {
			return err
		}
	}

	if err := n.disableScheduling(); err != nil {
		return err
	}

	if n.Isolate {
		if err := n.addTaint(); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) evictPods() error {

	var pods *corev1.PodList
	var err error

	listOpts := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + n.Name,
	}

	if pods, err = n.Flags.Client.CoreV1().Pods("").List(context.TODO(), listOpts); err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if pod.Spec.PriorityClassName != "system-node-critical" && !podIsInQuarantine(pod) && n.Isolate {
			if err := evictPod(pod, n.Flags.Client); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) update() error {

	ok, err := n.isAlreadyIsolated()

	if err != nil {
		n.Logger.Info(err.Error())
		return err
	}

	if err := n.manageWorkloads(); err != nil {
		return err
	}

	if !ok {
		n.Logger.Info("node not isolated...")

		n.Logger.Info("deschedule pods...")
		if err := n.deschedulePods(); err != nil {
			return err
		}

		n.Logger.Info("evict daemonset pods...")
		if err := n.evictPods(); err != nil {
			return err
		}
	}

	return nil
}

func (n Node) remove() error {
	return nil
}

func (n *Node) setNodesResources(rs []v1alpha1.Resource) {

	for _, r := range rs {
		switch t := r.Type; t {
		case dsType:
			n.Daemonsets = append(n.Daemonsets, Daemonset{
				Name:      r.Name,
				Namespace: r.Namespace,
				Keep:      r.Keep,
			})
		case deploymentType:
			n.Deployments = append(n.Deployments, Deployment{
				Name:      r.Name,
				Namespace: r.Namespace,
				Keep:      r.Keep,
			})
		}
	}
}

func (n *Node) mergeResources(rs []v1alpha1.Resource) {

	for _, r := range rs {
		switch t := r.Type; t {
		case dsType:
			if len(n.Daemonsets) > 1 {
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
				return
			}
			n.Daemonsets = append(n.Daemonsets, Daemonset{
				Name:      r.Name,
				Namespace: r.Namespace,
				Keep:      r.Keep,
			})
		case deploymentType:
			if len(n.Deployments) > 1 {
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
				return
			}
			n.Deployments = append(n.Deployments, Deployment{
				Name:      r.Name,
				Namespace: r.Namespace,
				Keep:      r.Keep,
			})
		}
	}
}

func (n *Node) parseFlags(baseFlags, nodeFlags v1alpha1.Flags) {

	falseFlag := false
	trueFlag := true

	flags := v1alpha1.Flags{
		DeleteEmptyDirData:  &trueFlag,
		DisableEviction:     &falseFlag,
		Force:               &falseFlag,
		IgnoreAllDaemonSets: &falseFlag,
		IgnoreErrors:        &falseFlag,
	}

	flags = n.mergeFlags(flags, baseFlags)
	flags = n.mergeFlags(flags, nodeFlags)

	n.Flags.DeleteEmptyDirData = *flags.DeleteEmptyDirData
	n.Flags.DisableEviction = *flags.DisableEviction
	n.Flags.Force = *flags.Force
	n.Flags.IgnoreAllDaemonSets = *flags.IgnoreAllDaemonSets
	n.Flags.IgnoreErrors = *flags.IgnoreErrors
}

func (n *Node) mergeFlags(baseFlags, mergeFlags v1alpha1.Flags) v1alpha1.Flags {

	if mergeFlags.DeleteEmptyDirData != nil {
		baseFlags.DeleteEmptyDirData = mergeFlags.DeleteEmptyDirData
	}

	if mergeFlags.DisableEviction != nil {
		baseFlags.DisableEviction = mergeFlags.DisableEviction
	}

	if mergeFlags.Force != nil {
		baseFlags.Force = mergeFlags.Force
	}

	if mergeFlags.IgnoreAllDaemonSets != nil {
		baseFlags.IgnoreAllDaemonSets = mergeFlags.IgnoreAllDaemonSets
	}

	if mergeFlags.IgnoreErrors != nil {
		baseFlags.IgnoreErrors = mergeFlags.IgnoreErrors
	}

	return baseFlags
}

func (n Node) disableScheduling() error {

	nodeObj := n.getNodeAPIObject()

	n.Logger.Info("cordon...")

	if err := drain.RunCordonOrUncordon(n.Flags, nodeObj, true); err != nil {
		return err
	}

	timeout := int64(20)
	listOpts := metav1.ListOptions{
		Watch:          true,
		LabelSelector:  "kubernetes.io/hostname=" + n.Name,
		TimeoutSeconds: &timeout,
	}
	w, err := n.Flags.Client.CoreV1().Nodes().Watch(context.TODO(), listOpts)

	if err != nil {
		return err
	}

	waitForResource(w, n.Logger)

	return nil
}

func (n Node) enableScheduling() error {

	nodeObj := n.getNodeAPIObject()

	n.Logger.Info("uncordon...")

	if err := drain.RunCordonOrUncordon(n.Flags, nodeObj, false); err != nil {
		return err
	}

	if err := n.waitForUpdate(); err != nil {
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

	if err := n.waitForUpdate(); err != nil {
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

	if err := n.waitForUpdate(); err != nil {
		return err
	}

	return nil
}

func (n Node) deschedulePods() error {
	if err := drain.RunNodeDrain(n.Flags, n.Name); err != nil {
		return err
	}

	if err := n.waitForUpdate(); err != nil {
		return err
	}

	return nil
}

func (n Node) getNodeAPIObject() *corev1.Node {

	var err error
	var nodeObj *corev1.Node

	opts := metav1.GetOptions{}
	core := n.Flags.Client.CoreV1()
	if nodeObj, err = core.Nodes().Get(context.Background(), n.Name, opts); err != nil {
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

	return nil
}

func (n Node) waitForUpdate() error {
	timeout := int64(20)
	listOpts := metav1.ListOptions{
		Watch:          true,
		LabelSelector:  "kubernetes.io/hostname=" + n.Name,
		TimeoutSeconds: &timeout,
	}

	apps := n.Flags.Client.AppsV1()
	_ = apps.Deployments("")

	w, err := n.Flags.Client.CoreV1().Nodes().Watch(context.TODO(), listOpts)

	if err != nil {
		return err
	}

	waitForResource(w, n.Logger)

	return nil
}

func (n Node) isAlreadyIsolated() (bool, error) {

	nodeObj := n.getNodeAPIObject()
	return nodeObj.Spec.Unschedulable, nil
}
