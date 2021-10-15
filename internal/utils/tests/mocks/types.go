package mocks

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
)

type K8SFactoryMock struct {
	mock.Mock
	util.Factory
}

type K8SClientsetMock struct {
	mock.Mock
	kubernetes.Clientset
}
