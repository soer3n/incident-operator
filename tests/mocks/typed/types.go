package typed

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	policyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
)

// Client represents mock struct for k8s runtime client
type Client struct {
	mock.Mock
	kubernetes.Interface
}

// CoreV1 represents mock struct for k8s runtime client core v1 api resources
type Discovery struct {
	mock.Mock
	discovery.DiscoveryInterface
}

// CoreV1 represents mock struct for k8s runtime client core v1 api resources
type CoreV1 struct {
	mock.Mock
	corev1.CoreV1Interface
}

// AppsV1 represents mock struct for k8s runtime client apps v1 api resources
type AppsV1 struct {
	mock.Mock
	appsv1.AppsV1Interface
}

type PolicyV1Beta1 struct {
	mock.Mock
	policyv1beta1.PolicyV1beta1Interface
}

// NodeV1 represents mock struct for k8s runtime client  v1 node resources
type NodeV1 struct {
	mock.Mock
	corev1.NodeInterface
}

// PodV1 represents mock struct for k8s runtime client v1 pod resources
type PodV1 struct {
	mock.Mock
	corev1.PodInterface
}

// DeploymentV1 represents mock struct for k8s runtime client v1 deployment resources
type DeploymentV1 struct {
	mock.Mock
	appsv1.DeploymentInterface
}

// DaemonsetV1 represents mock struct for k8s runtime client v1 deployment resources
type DaemonsetV1 struct {
	mock.Mock
	appsv1.DaemonSetInterface
}
