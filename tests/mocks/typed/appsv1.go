package typed

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// AppsV1 represents mock func for similar runtime client func
func (client *Client) AppsV1() appsv1.AppsV1Interface {
	args := client.Called()
	v := args.Get(0)
	return v.(appsv1.AppsV1Interface)
}

// Deployments represents mock func for similar runtime client func
func (a *AppsV1) Deployments(namespace string) appsv1.DeploymentInterface {
	args := a.Called(namespace)
	v := args.Get(0)
	return v.(appsv1.DeploymentInterface)
}

// Daemonsets represents mock func for similar runtime client func
func (a *AppsV1) DaemonSets(namespace string) appsv1.DaemonSetInterface {
	args := a.Called(namespace)
	v := args.Get(0)
	return v.(appsv1.DaemonSetInterface)
}

// Get represents mock func for similar runtime client func
func (getter *DeploymentV1) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Deployment, error) {
	args := getter.Called(ctx, name, options)
	values := args.Get(0).(*v1.Deployment)
	err := args.Error(1)
	return values, err
}

// List represents mock func for similar runtime client func
func (getter *DeploymentV1) List(ctx context.Context, opts metav1.ListOptions) (*v1.DeploymentList, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(*v1.DeploymentList)
	err := args.Error(1)
	return values, err
}

// Watch represents mock func for similar runtime client func
func (getter *DeploymentV1) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := getter.Called(opts)
	values := args.Get(0).(watch.Interface)
	err := args.Error(1)
	return values, err
}

// Create represents mock func for similar runtime client func
func (getter *DeploymentV1) Create(ctx context.Context, obj *v1.Deployment, options metav1.CreateOptions) (*v1.Deployment, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Deployment)
	err := args.Error(1)
	return values, err
}

// Patch represents mock func for similar runtime client func
func (getter *DeploymentV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*v1.Deployment, error) {
	args := getter.Called(ctx, options, pt, data, options, subresources)
	values := args.Get(0).(*v1.Deployment)
	err := args.Error(1)
	return values, err
}

// UpdateStatus represents mock func for similar runtime client func
func (getter *DeploymentV1) UpdateStatus(ctx context.Context, obj *v1.Deployment, options metav1.UpdateOptions) (*v1.Deployment, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Deployment)
	err := args.Error(1)
	return values, err
}

// Update represents mock func for similar runtime client func
func (getter *DeploymentV1) Update(ctx context.Context, obj *v1.Deployment, options metav1.UpdateOptions) (*v1.Deployment, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.Deployment)
	err := args.Error(1)
	return values, err
}

// Delete represents mock func for similar runtime client func
func (getter *DeploymentV1) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	args := getter.Called(ctx, name, options)
	err := args.Error(0)
	return err
}

// DeleteCollection represents mock func for similar runtime client func
func (getter *DeploymentV1) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	args := getter.Called(ctx, options, listOptions)
	err := args.Error(0)
	return err
}

// Get represents mock func for similar runtime client func
func (getter *DaemonsetV1) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.DaemonSet, error) {
	args := getter.Called(ctx, name, options)
	values := args.Get(0).(*v1.DaemonSet)
	err := args.Error(1)
	return values, err
}

// List represents mock func for similar runtime client func
func (getter *DaemonsetV1) List(ctx context.Context, opts metav1.ListOptions) (*v1.DaemonSetList, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(*v1.DaemonSetList)
	err := args.Error(1)
	return values, err
}

// Watch represents mock func for similar runtime client func
func (getter *DaemonsetV1) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := getter.Called(ctx, opts)
	values := args.Get(0).(watch.Interface)
	err := args.Error(1)
	return values, err
}

// Create represents mock func for similar runtime client func
func (getter *DaemonsetV1) Create(ctx context.Context, obj *v1.DaemonSet, options metav1.CreateOptions) (*v1.DaemonSet, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.DaemonSet)
	err := args.Error(1)
	return values, err
}

// Patch represents mock func for similar runtime client func
func (getter *DaemonsetV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*v1.DaemonSet, error) {
	args := getter.Called(ctx, name, pt, data, options, subresources)
	values := args.Get(0).(*v1.DaemonSet)
	err := args.Error(1)
	return values, err
}

// UpdateStatus represents mock func for similar runtime client func
func (getter *DaemonsetV1) UpdateStatus(ctx context.Context, obj *v1.DaemonSet, options metav1.UpdateOptions) (*v1.DaemonSet, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.DaemonSet)
	err := args.Error(1)
	return values, err
}

// Update represents mock func for similar runtime client func
func (getter *DaemonsetV1) Update(ctx context.Context, obj *v1.DaemonSet, options metav1.UpdateOptions) (*v1.DaemonSet, error) {
	args := getter.Called(ctx, obj, options)
	values := args.Get(0).(*v1.DaemonSet)
	err := args.Error(1)
	return values, err
}

// Delete represents mock func for similar runtime client func
func (getter *DaemonsetV1) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	args := getter.Called(ctx, name, options)
	err := args.Error(0)
	return err
}

// DeleteCollection represents mock func for similar runtime client func
func (getter *DaemonsetV1) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	args := getter.Called(ctx, options, listOptions)
	err := args.Error(0)
	return err
}
