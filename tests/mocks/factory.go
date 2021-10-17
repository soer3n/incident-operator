package mocks

import "k8s.io/client-go/kubernetes"

// KuerbetesClientSet represents mock func for similar for getting clientset
func (c *K8SFactoryMock) KubernetesClientSet() (*kubernetes.Clientset, error) {
	args := c.Called()
	kc := args.Get(0).(*kubernetes.Clientset)
	err := args.Error(1)
	return kc, err
}
