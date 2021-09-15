package quarantine

import (
	"errors"
	"os"

	"github.com/prometheus/common/log"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/soer3n/incident-operator/api/v1alpha1"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const quarantinePodSelector = "quarantine"
const quarantineTaintKey = quarantinePodSelector
const quarantineTaintValue = "true"
const quarantineTaintEffect = "NoExecute"
const quarantineStatusActiveKey = "active"
const quarantineStatusActiveMessage = "success"

// New represents an initialization of a quarantine struct
func New(s *v1alpha1.Quarantine) (*Quarantine, error) {

	debugImage := debugPodImage
	debugNamespace := debugPodNamespace

	if s.Spec.Debug.Image != "" {
		debugImage = s.Spec.Debug.Image
	}

	if s.Spec.Debug.Namespace != "" {
		debugNamespace = s.Spec.Debug.Namespace
	}

	q := &Quarantine{
		Debug: Debug{
			Enabled:   s.Spec.Debug.Enabled,
			Image:     debugImage,
			Namespace: debugNamespace,
		},
		isActive:   false,
		conditions: s.Status.Conditions,
	}
	nodes := []*Node{}

	for _, n := range s.Spec.Nodes {
		temp := &Node{
			Name: n.Name,
			Debug: Debug{
				Enabled:   s.Spec.Debug.Enabled,
				Image:     debugImage,
				Namespace: debugNamespace,
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

		temp.parseFlags()
		nodes = append(nodes, temp)
	}

	q.Nodes = nodes

	if len(s.Status.Conditions) > 0 {
		q.isActive = true
	}

	return q, nil
}

// Prepare represents the tasks before a quarantine can be started
func (q *Quarantine) Prepare() error {

	for _, n := range q.Nodes {
		if q.Debug.Enabled {
			if err := q.Debug.deploy(n.flags.Client, n.Name); err != nil {
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

// Start represents the tasks to start isolating resources on nodes
func (q *Quarantine) Start() error {

	for _, n := range q.Nodes {
		if err := n.deschedulePods(); err != nil {
			return err
		}
	}

	return nil
}

// Update represents the tasks which are not yet executed
func (q *Quarantine) Update() error {

	// limit update to fix failed reconciles
	if meta.IsStatusConditionPresentAndEqual(q.conditions, quarantineStatusActiveKey, metav1.ConditionTrue) &&
		q.conditions[0].Message == quarantineStatusActiveMessage {
		return nil
	}

	for _, n := range q.Nodes {
		if err := n.update(); err != nil {
			return err
		}
	}

	return nil
}

// Stop represents the tasks for uncordon nodes, rescheduling resources and deleting debug resources
func (q *Quarantine) Stop() error {

	if len(q.Nodes) < 1 {
		return errors.New("no nodes detected")
	}

	if err := q.Debug.remove(q.Nodes[0].flags.Client, debugPodName, q.Debug.Namespace); err != nil {
		return err
	}

	for _, n := range q.Nodes {
		if err := n.removeTaint(); err != nil {
			return err
		}

		for _, ds := range n.Daemonsets {
			if err := ds.removeToleration(n.flags.Client); err != nil {
				return err
			}
		}

		if err := n.enableScheduling(); err != nil {
			return err
		}
	}

	if err := cleanupIsolatedPods(q.Nodes[0].flags.Client); err != nil {
		return err
	}

	return nil
}

// IsActive represents returning state of a quarantine
func (q Quarantine) IsActive() bool {
	return q.isActive
}
