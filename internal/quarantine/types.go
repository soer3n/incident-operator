package quarantine

import (
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/drain"
)

// Quarantine represents current state of isolation
type Quarantine struct {
	Nodes      []*Node
	Debug      Debug
	Client     kubernetes.Interface
	isActive   bool
	Conditions []metav1.Condition
	Logger     logr.Logger
}

// Node represents configuration for isolating a node
type Node struct {
	Name        string
	Debug       Debug
	Isolate     bool
	Daemonsets  []Daemonset
	Deployments []Deployment
	IOStreams   genericclioptions.IOStreams
	factory     util.Factory
	Flags       *drain.Helper
	Logger      logr.Logger
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
	Keep      bool
}

// Daemonset represents a configuration for a daemonset whose pod which is on an affected node should be isolated
type Daemonset struct {
	Name      string
	Namespace string
	Keep      bool
}
