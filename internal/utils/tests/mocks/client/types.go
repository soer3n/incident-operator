package client

import (
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8SClientMock represents mock struct for k8s runtime client
type K8SClientMock struct {
	mock.Mock
	client.Client
}
