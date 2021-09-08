package quarantine

import (
	"os"

	"github.com/prometheus/common/log"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/soer3n/incident-operator/api/v1alpha1"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const QuarantineLabelSelector = "quarantine"
const QuarantineTaintKey = QuarantineLabelSelector
const QuarantineTaintValue = "true"
const QuarantineStatusActiveKey = "active"
const QuarantineStatusActiveMessage = "success"

func New(s *v1alpha1.Quarantine) (*Quarantine, error) {

	q := &Quarantine{
		Debug: Debug{
			Enabled: s.Spec.Debug,
		},
		isActive: false,
	}
	nodes := []*Node{}

	for _, n := range s.Spec.Nodes {
		temp := &Node{
			Name: n.Name,
			Debug: Debug{
				Enabled: s.Spec.Debug,
			},
			isolate: n.Isolate,
			ioStreams: genericclioptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stdout,
			},
			factory: util.NewFactory(genericclioptions.NewConfigFlags(false)),
		}

		if err := temp.mergeResources(s.Spec.Resources); err != nil {
			return q, err
		}

		nodes = append(nodes, temp)
	}

	q.Nodes = nodes

	if meta.IsStatusConditionPresentAndEqual(s.Status.Conditions, QuarantineStatusActiveKey, metav1.ConditionTrue) &&
		s.Status.Conditions[0].Message == QuarantineStatusActiveMessage {
		q.isActive = true
	}

	return q, nil
}

func (q *Quarantine) Prepare() error {

	for _, n := range q.Nodes {
		if q.Debug.Enabled {
			if err := q.Debug.deploy(n.Name); err != nil {
				return err
			}
		}

		if ok, err := n.isAlreadyIsolated(); !ok {
			if err != nil {
				log.Info(err.Error())
			}

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

		if n.isolate {
			if err := n.addTaint(); err != nil {
				return err
			}
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

			return nil
		}
	}

	return nil
}

func (q *Quarantine) Stop() error {
	return nil
}

func (q Quarantine) IsActive() bool {
	return q.isActive
}
