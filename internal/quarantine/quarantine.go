package quarantine

import (
	"github.com/soer3n/incident-operator/api/v1alpha1"
)

func New(s *v1alpha1.Quarantine) (*Quarantine, error) {

	nodes := []*Node{}

	for _, n := range s.Spec.Nodes {
		temp := &Node{
			Name: n.Name,
			Debug: Debug{
				Enabled: s.Spec.Debug,
			},
		}

		if err := temp.mergeResources(s.Spec.Resources); err != nil {
			return &Quarantine{}, err
		}
	}

	return &Quarantine{
		Nodes: nodes,
		Debug: Debug{
			Enabled: s.Spec.Debug,
		},
	}, nil
}

func (q *Quarantine) Prepare() error {

	for _, n := range q.Nodes {
		if q.Debug.Enabled {
			if err := q.Debug.deploy(n.Name); err != nil {
				return err
			}
		}

		if !n.isAlreadyIsolated() {
			if err := n.prepare(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (q *Quarantine) Start() error {

	for _, n := range q.Nodes {
		if err := n.deschedulePods(); err != nil {
			return err
		}
	}

	return nil
}

func (q *Quarantine) Update() error {

	for _, n := range q.Nodes {
		if n.daemonsetsNotEqual() || n.deploymentsNotEqual() {
			if err := n.update(); err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}

func (q *Quarantine) Stop() error {
	return nil
}

func (q Quarantine) IsActive() bool {
	return false
}
