package quarantine

import (
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/yaho/pkg/client"
)

const DsType = "daemonset"
const DeploymentType = "deployment"

func (n Node) prepare() error {

	if err := n.disableScheduling(); err != nil {
		return err
	}

	for _, ds := range n.Daemonsets {

		if err := ds.isolatePod(); err != nil {
			return err
		}
	}

	for _, d := range n.Deployments {

		if err := d.isolatePod(); err != nil {
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
	return nil
}

func (n Node) deschedulePods() error {
	return nil
}

func (n Node) isAlreadyIsolated() bool {
	_ = client.New()
	return false
}

func (n Node) deploymentsNotEqual() bool {
	return false
}

func (n Node) daemonsetsNotEqual() bool {
	return false
}
