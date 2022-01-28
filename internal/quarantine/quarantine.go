package quarantine

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/go-logr/logr"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/drain"

	"github.com/soer3n/incident-operator/api/v1alpha1"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const quarantinePodSelector = "quarantine"
const quarantineTaintKey = quarantinePodSelector
const quarantineTaintValue = "true"
const quarantineTaintOperator = "Exists"
const quarantineTaintEffect = "NoSchedule"
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
		Conditions: s.Status.Conditions,
		Logger:     reqLogger,
	}
	nodes := []*Node{}

	for _, n := range s.Spec.Nodes {
		temp := q.getNodeStruct(n.Name, debugImage, debugNamespace, n.Isolate, f)
		temp.setNodeResources(n.Resources)
		temp.mergeResources(s.Spec.Resources)
		temp.parseFlags(s.Spec.Flags, n.Flags)
		nodes = append(nodes, temp)
	}

	q.Nodes = nodes

	nodesToRemove := []string{}
	nodesToRemoveObj := []*Node{}

	if _, ok := s.ObjectMeta.Annotations[QuarantinePodLabelPrefix+QuarantineNodeRemoveLabel]; ok {
		nodesToRemove = strings.Split(s.ObjectMeta.Annotations[QuarantinePodLabelPrefix+QuarantineNodeRemoveLabel], ",")
	}

	for _, r := range nodesToRemove {
		temp := q.getNodeStruct(r, debugImage, debugNamespace, false, f)
		nodesToRemoveObj = append(nodesToRemoveObj, temp)
	}

	q.MarkedNodes = nodesToRemoveObj

	if len(s.Status.Conditions) > 0 {
		q.isActive = true
	}

	return q, nil
}

func (q Quarantine) getNodeStruct(name, debugImage, debugNamespace string, isolate bool, f util.Factory) *Node {
	return &Node{
		Name:        name,
		Daemonsets:  []Daemonset{},
		Deployments: []Deployment{},
		Debug: Debug{
			Enabled:   q.Debug.Enabled,
			Image:     debugImage,
			Namespace: debugNamespace,
		},
		Isolate: isolate,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stdout,
		},
		factory: f,
		Logger:  q.Logger.WithValues("node", name),
		Flags: &drain.Helper{
			IgnoreAllDaemonSets: true,
			DisableEviction:     false,
			DeleteEmptyDirData:  true,
			PodSelector:         "!" + QuarantinePodLabelPrefix + quarantinePodSelector,
			Force:               false,
			IgnoreErrors:        false,
			Ctx:                 context.TODO(),
			Client:              q.Client,
			ErrOut:              os.Stdout,
			Out:                 os.Stdout,
		},
	}
}

// Prepare represents the tasks before a quarantine can be started
func (q *Quarantine) Prepare() error {

	for _, n := range q.Nodes {

		q.Logger.Info("preparing node...", "node", n.Name)

		if q.Debug.Enabled || n.Debug.Enabled {
			q.Logger.Info("deploying debug pod...", "node", n.Name)
			if err := q.Debug.deploy(n.Flags.Client, n.Name); err != nil {
				return err
			}
		}

		if ok, err := n.isAlreadyIsolated(); !ok {
			if err != nil {
				q.Logger.Info(err.Error())
			}

			q.Logger.Info("updating node...", "node", n.Name)
			if err := n.manageWorkloads(); err != nil {
				return err
			}
			continue
		}

		q.Logger.Info("already isolated...", "node", n.Name)
	}

	return nil
}

// Start represents the tasks to start isolating resources on nodes
func (q *Quarantine) Start() error {

	for _, n := range q.Nodes {

		q.Logger.Info("deschedule pods...", "node", n.Name)
		if err := n.deschedulePods(); err != nil {
			return err
		}

		n.Logger.Info("evict daemonset pods...", "node", n.Name)
		if err := n.evictPods(); err != nil {
			return err
		}
	}

	return nil
}

// Update represents the tasks which are not yet executed
func (q *Quarantine) Update() error {

	for _, n := range q.MarkedNodes {
		if q.Debug.Enabled || n.Debug.Enabled {
			q.Logger.Info("remove debug pods...")
			q.Debug.remove(q.Client, n.Name, q.Logger)
		}

		if err := n.remove(); err != nil {
			return err
		}
	}

	// limit update to fix failed reconciles
	if meta.IsStatusConditionPresentAndEqual(q.Conditions, quarantineStatusActiveKey, metav1.ConditionTrue) &&
		q.Conditions[0].Message == quarantineStatusActiveMessage {
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

		if q.Debug.Enabled || n.Debug.Enabled {
			q.Logger.Info("remove debug pods...")
			q.Debug.remove(q.Client, n.Name, q.Logger)
		}

		if err := n.remove(); err != nil {
			return err
		}

		for _, ds := range n.Daemonsets {
			q.Logger.Info("remove toleration for daemonset...", "dameonset", ds.Name)
			if err := ds.removeToleration(n.Flags.Client); err != nil {
				return err
			}
		}

		for _, d := range n.Deployments {
			q.Logger.Info("remove toleration for deployment...", "deployment", d.Name)
			if err := d.removeToleration(n.Flags.Client); err != nil {
				return err
			}
		}
	}

	q.Logger.Info("clean up isolated pods...")
	if err := cleanupIsolatedPods(q.Client); err != nil {
		return err
	}

	return nil
}

// IsActive represents returning state of a quarantine
func (q Quarantine) IsActive() bool {
	return q.isActive
}
