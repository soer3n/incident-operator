package quarantine

import "github.com/soer3n/incident-operator/api/v1alpha1"

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
	return nil
}

func (n Node) disableScheduling() error {
	return nil
}

func (n Node) deschedulePods() error {
	return nil
}

func (n Node) isAlreadyIsolated() bool {
	return false
}

func (n Node) deploymentsNotEqual() bool {
	return false
}

func (n Node) daemonsetsNotEqual() bool {
	return false
}
