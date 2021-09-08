package quarantine

import (
	"context"

	"k8s.io/kubectl/pkg/cmd/drain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/yaho/pkg/client"
)

const DsType = "daemonset"
const DeploymentType = "deployment"
const DrainIgnoreDaemonSetFlag = "--ignore-daemonsets"
const DrainPodSelectorFlag = "--pod-selector"
const DrainDisableEvictionFlag = "--disable-eviction"
const DrainDeleteEmptyDirDataFlag = "--delete-emptdir-data"

func (n Node) prepare() error {

	if err := n.disableScheduling(); err != nil {
		return err
	}

	for _, ds := range n.Daemonsets {

		if err := ds.isolatePod(client.New().TypedClient, n.isolate); err != nil {
			return err
		}
	}

	for _, d := range n.Deployments {

		if err := d.isolatePod(client.New().TypedClient); err != nil {
			return err
		}
	}

	return nil
}

func (n Node) update() error {
	return nil
}

func (n *Node) mergeResources(rs []v1alpha1.Resource) error {

	for _, r := range rs {
		switch t := r.Type; t {
		case DsType:
			for _, v := range n.Daemonsets {
				if v.Name == r.Name && v.Namespace == r.Namespace {
					continue
				}
				n.Daemonsets = append(n.Daemonsets, v)
			}
		case DeploymentType:
			for _, v := range n.Deployments {
				if v.Name == r.Name && v.Namespace == r.Namespace {
					continue
				}
				n.Deployments = append(n.Deployments, v)
			}
		}
	}

	return nil
}

func (n Node) disableScheduling() error {

	cordonOpts := drain.NewDrainCmdOptions(n.factory, n.ioStreams)
	cmd := drain.NewCmdCordon(n.factory, n.ioStreams)
	nodes := []string{
		n.Name,
	}

	if err := cordonOpts.Complete(n.factory, cmd, nodes); err != nil {
		return err
	}

	if err := cordonOpts.RunCordonOrUncordon(true); err != nil {
		return err
	}

	return nil
}

func (n Node) addTaint() error {
	return nil
}

func (n Node) deschedulePods() error {

	drainOpts := drain.NewDrainCmdOptions(n.factory, n.ioStreams)
	cmd := drain.NewCmdDrain(n.factory, n.ioStreams)
	args := []string{
		n.Name,
		DrainIgnoreDaemonSetFlag,
	}

	if err := drainOpts.Complete(n.factory, cmd, args); err != nil {
		return err
	}

	if err := drainOpts.RunDrain(); err != nil {
		return err
	}

	return nil
}

func (n Node) isAlreadyIsolated() (bool, error) {

	opts := metav1.GetOptions{}
	obj, err := client.New().TypedClient.CoreV1().Nodes().Get(context.Background(), n.Name, opts)

	if err != nil {
		return obj.Spec.Unschedulable, err
	}

	return obj.Spec.Unschedulable, nil
}

func (n Node) deploymentsNotEqual() bool {
	return false
}

func (n Node) daemonsetsNotEqual() bool {
	return false
}
