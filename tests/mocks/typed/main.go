package typed

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CoreV1 represents mock func for similar runtime client func
func (client *Client) CoreV1() corev1.CoreV1Interface {
	args := client.Called()
	v := args.Get(0)
	return v.(corev1.CoreV1Interface)
}

// AppsV1 represents mock func for similar runtime client func
func (client *Client) AppsV1() appsv1.AppsV1Interface {
	args := client.Called()
	v := args.Get(0)
	return v.(appsv1.AppsV1Interface)
}

// Discovery represents mock func for similar runtime client func
func (client *Client) Discovery() discovery.DiscoveryInterface {
	args := client.Called()
	v := args.Get(0)
	return v.(discovery.DiscoveryInterface)
}

// ServerGroups represents mock func for similar runtime client func
func (discovery *Discovery) ServerGroups() (*metav1.APIGroupList, error) {
	args := discovery.Called()
	v := args.Get(0).(*metav1.APIGroupList)
	err := args.Error(1)
	return v, err
}
