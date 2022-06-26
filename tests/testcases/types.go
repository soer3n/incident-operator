package testcases

import (
	mocks "github.com/soer3n/incident-operator/tests/mocks/typed"
	corev1 "k8s.io/api/core/v1"
)

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
	Pods        []*TestClientPod
}

type TestClientResource struct {
	Name          string
	Type          string
	Node          string
	Watch         bool
	Isolated      bool
	GracePeriod   bool
	Taint         bool
	Error         TestClientError
	ListSelector  []string
	FieldSelector []string
}

type TestClientPod struct {
	Resource TestClientResource
	pod      *corev1.Pod
}

type TestClientError struct {
	Enabled bool
	Message string
}

type TestClientSelectors struct {
	ListSelectors  map[string][]string
	FieldSelectors map[string][]string
}

type TestClientSelectorValues struct {
	Value string
	Pods  []corev1.Pod
}

type TestClientSelectorStruct struct {
	ListSelectors  map[string]TestClientSelectorValues
	FieldSelectors map[string]TestClientSelectorValues
}
