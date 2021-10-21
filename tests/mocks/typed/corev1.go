package typed

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CoreV1 represents mock func for similar runtime client func
func (client *Client) CoreV1() corev1.CoreV1Interface {
	args := client.Called()
	v := args.Get(0)
	return v.(corev1.CoreV1Interface)
}

// Nodes represents mock func for similar runtime client func
func (c *CoreV1) Nodes() corev1.NodeInterface {
	args := c.Called()
	v := args.Get(0)
	return v.(corev1.NodeInterface)
}

// Pods represents mock func for similar runtime client func
func (c *CoreV1) Pods(namespace string) corev1.PodInterface {
	args := c.Called(namespace)
	v := args.Get(0)
	return v.(corev1.PodInterface)
}

// Get represents mock func for similar runtime client func
func (getter *NodeV1) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Node, error) {
	args := getter.Called(ctx, name, options)
	values := args.Get(0).(*v1.Node)
	err := args.Error(1)
	return values, err
}

// List represents mock func for similar runtime client func
func (getter *NodeV1) List(ctx context.Context, opts metav1.ListOptions) (*v1.NodeList, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(*v1.NodeList)
	err := args.Error(1)
	return values, err
}

// Watch represents mock func for similar runtime client func
func (getter *NodeV1) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(watch.Interface)
	err := args.Error(1)
	return values, err
}

// Create represents mock func for similar runtime client func
func (getter *NodeV1) Create(ctx context.Context, obj *v1.Node, options metav1.CreateOptions) (*v1.Node, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Node)
	err := args.Error(1)
	return values, err
}

// Patch represents mock func for similar runtime client func
func (getter *NodeV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*v1.Node, error) {
	args := getter.Called(ctx, name, pt, data, options, subresources)
	values := args.Get(0).(*v1.Node)
	err := args.Error(1)
	return values, err
}

// UpdateStatus represents mock func for similar runtime client func
func (getter *NodeV1) UpdateStatus(ctx context.Context, obj *v1.Node, options metav1.UpdateOptions) (*v1.Node, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Node)
	err := args.Error(1)
	return values, err
}

// Update represents mock func for similar runtime client func
func (getter *NodeV1) Update(ctx context.Context, obj *v1.Node, options metav1.UpdateOptions) (*v1.Node, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Node)
	err := args.Error(1)
	return values, err
}

// Delete represents mock func for similar runtime client func
func (getter *NodeV1) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	args := getter.Called(ctx, name, options)
	err := args.Error(0)
	return err
}

// DeleteCollection represents mock func for similar runtime client func
func (getter *NodeV1) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	args := getter.Called(ctx, options, listOptions)
	err := args.Error(0)
	return err
}

// Get represents mock func for similar runtime client func
func (getter *PodV1) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Pod, error) {
	args := getter.Called(ctx, name, options)
	values := args.Get(0).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// List represents mock func for similar runtime client func
func (getter *PodV1) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(*v1.PodList)
	err := args.Error(1)
	return values, err
}

// Watch represents mock func for similar runtime client func
func (getter *PodV1) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(watch.Interface)
	err := args.Error(1)
	return values, err
}

// Create represents mock func for similar runtime client func
func (getter *PodV1) Create(ctx context.Context, obj *v1.Pod, options metav1.CreateOptions) (*v1.Pod, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(1).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// Patch represents mock func for similar runtime client func
func (getter *PodV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*v1.Pod, error) {
	args := getter.Called(ctx, name, pt, data, options, subresources)
	values := args.Get(0).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// UpdateStatus represents mock func for similar runtime client func
func (getter *PodV1) UpdateStatus(ctx context.Context, obj *v1.Pod, options metav1.UpdateOptions) (*v1.Pod, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// Update represents mock func for similar runtime client func
func (getter *PodV1) Update(ctx context.Context, obj *v1.Pod, options metav1.UpdateOptions) (*v1.Pod, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// Delete represents mock func for similar runtime client func
func (getter *PodV1) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	args := getter.Called(ctx, name, options)
	err := args.Error(0)
	return err
}

// DeleteCollection represents mock func for similar runtime client func
func (getter *PodV1) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	args := getter.Called(ctx, options, listOptions)
	err := args.Error(0)
	return err
}

// Apply represents mock func for similar runtime client func
func (getter *PodV1) Apply(ctx context.Context, pod *applyv1.PodApplyConfiguration, opts metav1.ApplyOptions) (*v1.Pod, error) {
	args := getter.Called(ctx, pod, opts)
	values := args.Get(0).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// ApplyStatus represents mock func for similar runtime client func
func (getter *PodV1) ApplyStatus(ctx context.Context, pod *applyv1.PodApplyConfiguration, opts metav1.ApplyOptions) (*v1.Pod, error) {
	args := getter.Called(ctx, pod, opts)
	values := args.Get(0).(*v1.Pod)
	err := args.Error(1)
	return values, err
}

// GetEphemeralContainers represents mock func for similar runtime client func
func (getter *PodV1) GetEphemeralContainers(ctx context.Context, podName string, options metav1.GetOptions) (*v1.EphemeralContainers, error) {
	args := getter.Called(ctx, podName, options)
	values := args.Get(0).(*v1.EphemeralContainers)
	err := args.Error(1)
	return values, err
}

// UpdateEphemeralContainers represents mock func for similar runtime client func
func (getter *PodV1) UpdateEphemeralContainers(ctx context.Context, podName string, ephemeralContainers *v1.EphemeralContainers, opts metav1.UpdateOptions) (*v1.EphemeralContainers, error) {
	args := getter.Called(ctx, podName, ephemeralContainers, opts)
	values := args.Get(0).(*v1.EphemeralContainers)
	err := args.Error(1)
	return values, err
}
