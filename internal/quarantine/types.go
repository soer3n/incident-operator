package quarantine

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/drain"
)

type Quarantine struct {
	Nodes      []*Node
	Debug      Debug
	isActive   bool
	conditions []metav1.Condition
}

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

type Debug struct {
	Image     string
	Namespace string
	Enabled   bool
}

type Deployment struct {
	Name      string
	Namespace string
}

type Daemonset struct {
	Name      string
	Namespace string
}
