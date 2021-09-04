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

type Deployment struct {
	Name      string
	Namespace string
}

type Daemonset struct {
	Name      string
	Namespace string
}
