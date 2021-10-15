package client

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// List represents mock func for similar dynamic runtime client func
func (client *K8SClientMock) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := client.Called(ctx, list, opts)
	err := args.Error(0)
	return err
}

// Get represents mock func for similar dynamic runtime client func
func (client *K8SClientMock) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	args := client.Called(ctx, key, obj)
	err := args.Error(0)
	return err
}
