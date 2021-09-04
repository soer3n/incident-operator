package quarantine

type Quarantine struct {
	Nodes []*Node
	Debug Debug
}

type Node struct {
	Name        string
	Debug       Debug
	Daemonsets  []Daemonset
	Deployments []Deployment
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
