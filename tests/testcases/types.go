package testcases

import mocks "github.com/soer3n/incident-operator/tests/mocks/typed"

type TestClientQuarantine struct {
	FakeClient  *mocks.Client
	Nodes       []TestClientNode
	MarkedNodes []TestClientNode
	Namespaces  []TestClientNamespace
}

type TestClientNode struct {
	Name  string
	Drain bool
	Taint bool
}

type TestClientNamespace struct {
	Name        string
	Daemonsets  []TestClientResource
	Deployments []TestClientResource
	Pods        []TestClientResource
}

type TestClientResource struct {
	Name          string
	Type          string
	Node          string
	Watch         bool
	Isolated      bool
	Taint         bool
	Error         TestClientError
	ListSelector  []string
	FieldSelector []string
}

type TestClientError struct {
	Enabled bool
	Message string
}
