package quarantine

import (
	"errors"
	"os"

	"github.com/go-logr/logr"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
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
func New(s *v1alpha1.Quarantine, c kubernetes.Interface, f util.Factory, reqLogger logr.Logger) (*Quarantine, error) {

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
		Client:     c,
		isActive:   false,
		conditions: s.Status.Conditions,
		logger:     reqLogger,
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
			factory: f,
			logger:  reqLogger,
		}

		if err := temp.mergeResources(s.Spec.Resources); err != nil {
			return q, err
		}

		temp.parseFlags(c)
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

		q.logger.Info("preparing node...", "node", n.Name)

		if q.Debug.Enabled {
			q.logger.Info("deploying debug pod...", "node", n.Name)
			if err := q.Debug.deploy(n.Flags.Client, n.Name); err != nil {
				return err
			}
		}

		if ok, err := n.isAlreadyIsolated(); !ok {
			if err != nil {
				q.logger.Info(err.Error())
			}

			q.logger.Info("updating node...", "node", n.Name)
			if err := n.prepare(); err != nil {
				return err
			}
			continue
		}

		q.logger.Info("already isolated...", "node", n.Name)
	}

	return nil
}

// Start represents the tasks to start isolating resources on nodes
func (q *Quarantine) Start() error {

	for _, n := range q.Nodes {

		q.logger.Info("deschedule pods...", "node", n.Name)
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

	for _, n := range q.Nodes {

		if q.Debug.Enabled {
			q.logger.Info("remove debug pods...")
			q.Debug.remove(q.Nodes[0].Flags.Client, n.Name, q.logger)
		}

		q.logger.Info("remove taint...", "node", n.Name)
		if err := n.removeTaint(); err != nil {
			return err
		}

		for _, ds := range n.Daemonsets {
			q.logger.Info("remove toleration for daemonset...", "dameonset", ds.Name)
			if err := ds.removeToleration(n.Flags.Client); err != nil {
				return err
			}
		}

		q.logger.Info("enable scheduling again...", "node", n.Name)
		if err := n.enableScheduling(); err != nil {
			return err
		}
	}

	q.logger.Info("clean up isolated pods...")
	if err := cleanupIsolatedPods(q.Nodes[0].Flags.Client); err != nil {
		return err
	}

	return nil
}

// IsActive represents returning state of a quarantine
func (q Quarantine) IsActive() bool {
	return q.isActive
}
