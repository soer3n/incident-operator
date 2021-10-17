package tests

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/soer3n/incident-operator/internal/quarantine"
	mocks "github.com/soer3n/incident-operator/tests/mocks"
	"github.com/soer3n/incident-operator/tests/testcases"
	"github.com/stretchr/testify/assert"
)

func TestInitQuarantine(t *testing.T) {

	factoryMock := &mocks.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)
	quarantineSpecs := testcases.GetQuarantineInitSpec()
	logger := ctrl.Log.WithName("test")

	assert := assert.New(t)

	for _, spec := range quarantineSpecs {

		quarantine, err := quarantine.New(spec, fake.NewSimpleClientset(), factoryMock, logger)
		assert.Nil(err)
		assert.NotNil(quarantine)
	}
}

func TestStartQuarantine(t *testing.T) {

	quarantines := testcases.GetQuarantineStartStructs()

	for _, obj := range quarantines {

		err := obj.Input.Start()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestPrepareQuarantine(t *testing.T) {

	quarantines := testcases.GetQuarantinePrepareStructs()

	for _, obj := range quarantines {

		err := obj.Input.Prepare()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestStopQuarantine(t *testing.T) {

	quarantines := testcases.GetQuarantineStopStructs()

	for _, obj := range quarantines {

		err := obj.Input.Stop()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestUpdateQuarantine(t *testing.T) {

	quarantines := testcases.GetQuarantineStopStructs()

	for _, obj := range quarantines {

		err := obj.Input.Update()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestIsQuarantineActive(t *testing.T) {

	factoryMock := &mocks.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantines := testcases.GetQuarantineIsActiveStructs()

	for _, obj := range quarantines {

		isActive := obj.Input.IsActive()
		assert := assert.New(t)
		assert.False(isActive)
	}
}
