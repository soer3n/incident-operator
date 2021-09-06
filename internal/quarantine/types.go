package quarantine

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"
)

type Quarantine struct {
	Nodes    []*Node
	Debug    Debug
	isActive bool
}

type Node struct {
	Name        string
	Debug       Debug
	Daemonsets  []Daemonset
	Deployments []Deployment
	ioStreams   genericclioptions.IOStreams
	factory     util.Factory
}

type Debug struct {
	Image   string
	Enabled bool
}

type resource struct {
	Name      string
	Namespace string
}

type Deployment struct {
	resource
}

type Daemonset struct {
	resource
}
