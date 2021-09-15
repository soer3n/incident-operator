package quarantine

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/drain"
)

// Quarantine represents current state of isolation
type Quarantine struct {
	Nodes      []*Node
	Debug      Debug
	isActive   bool
	conditions []metav1.Condition
}

// Node represents configuration for isolating a node
type Node struct {
	Name        string
	Debug       Debug
	isolate     bool
	Daemonsets  []Daemonset
	Deployments []Deployment
	ioStreams   genericclioptions.IOStreams
	factory     util.Factory
	flags       *drain.Helper
}

// Debug represents a configuration for a debug pod
type Debug struct {
	Image     string
	Namespace string
	Enabled   bool
}

// Deployment represents a configuration for a deployment whose pod which is on an affected node should be isolated
type Deployment struct {
	Name      string
	Namespace string
}

// Daemonset represents a configuration for a daemonset whose pod which is on an affected node should be isolated
type Daemonset struct {
	Name      string
	Namespace string
}
