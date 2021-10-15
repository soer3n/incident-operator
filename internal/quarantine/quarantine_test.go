package quarantine

import (
	"os"
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubectl/pkg/drain"
	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	tmock "github.com/soer3n/incident-operator/internal/utils/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestInitQuarantine(t *testing.T) {

	factoryMock := &tmock.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)
	quarantineSpec := &v1alpha1.Quarantine{
		Spec: v1alpha1.QuarantineSpec{
			Debug: v1alpha1.Debug{
				Image:     "foo:bar",
				Namespace: "test",
			},
			Nodes: []v1alpha1.Node{
				{
					Name:    "worker1",
					Isolate: true,
				},
				{
					Name:    "worker2",
					Isolate: true,
				},
			},
		},
	}
	logger := ctrl.Log.WithName("test")

	quarantine, err := New(quarantineSpec, fake.NewSimpleClientset(), factoryMock, logger)

	assert := assert.New(t)
	assert.Nil(err)
	assert.NotNil(quarantine)
}

func TestStartQuarantine(t *testing.T) {

	factoryMock := &tmock.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantine := &Quarantine{
		Nodes: []*Node{
			{
				Name:        "foo",
				isolate:     false,
				Daemonsets:  []Daemonset{},
				Deployments: []Deployment{},
				logger:      ctrl.Log.WithName("test"),
				Flags:       &drain.Helper{},
				ioStreams: genericclioptions.IOStreams{
					In:     os.Stdin,
					Out:    os.Stdout,
					ErrOut: os.Stdout,
				},
				Debug: Debug{
					Enabled: false,
				},
			},
		},
		logger: ctrl.Log.WithName("test"),
		Debug: Debug{
			Enabled: false,
		},
		isActive:   false,
		conditions: []metav1.Condition{},
	}

	err := quarantine.Start()

	assert := assert.New(t)
	assert.Nil(err)
}

func TestPrepareQuarantine(t *testing.T) {

	factoryMock := &tmock.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantine := &Quarantine{
		Nodes: []*Node{
			{
				Name:        "foo",
				isolate:     false,
				Daemonsets:  []Daemonset{},
				Deployments: []Deployment{},
				logger:      ctrl.Log.WithName("test"),
				Flags:       &drain.Helper{},
				ioStreams: genericclioptions.IOStreams{
					In:     os.Stdin,
					Out:    os.Stdout,
					ErrOut: os.Stdout,
				},
				Debug: Debug{
					Enabled: false,
				},
			},
		},
		logger: ctrl.Log.WithName("test"),
		Debug: Debug{
			Enabled: false,
		},
		isActive:   false,
		conditions: []metav1.Condition{},
	}

	err := quarantine.Prepare()

	assert := assert.New(t)
	assert.Nil(err)
}

func TestStopQuarantine(t *testing.T) {

	factoryMock := &tmock.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantine := &Quarantine{
		Nodes: []*Node{
			{
				Name:        "foo",
				isolate:     false,
				Daemonsets:  []Daemonset{},
				Deployments: []Deployment{},
				logger:      ctrl.Log.WithName("test"),
				Flags:       &drain.Helper{},
				ioStreams: genericclioptions.IOStreams{
					In:     os.Stdin,
					Out:    os.Stdout,
					ErrOut: os.Stdout,
				},
				Debug: Debug{
					Enabled: false,
				},
			},
		},
		logger: ctrl.Log.WithName("test"),
		Debug: Debug{
			Enabled: false,
		},
		isActive:   false,
		conditions: []metav1.Condition{},
	}

	err := quarantine.Stop()

	assert := assert.New(t)
	assert.Nil(err)
}

func TestUpdateQuarantine(t *testing.T) {

	factoryMock := &tmock.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantine := &Quarantine{
		Nodes: []*Node{
			{
				Name:        "foo",
				isolate:     false,
				Daemonsets:  []Daemonset{},
				Deployments: []Deployment{},
				logger:      ctrl.Log.WithName("test"),
				Flags:       &drain.Helper{},
				ioStreams: genericclioptions.IOStreams{
					In:     os.Stdin,
					Out:    os.Stdout,
					ErrOut: os.Stdout,
				},
				Debug: Debug{
					Enabled: false,
				},
			},
		},
		logger: ctrl.Log.WithName("test"),
		Debug: Debug{
			Enabled: false,
		},
		isActive:   false,
		conditions: []metav1.Condition{},
	}

	err := quarantine.Update()

	assert := assert.New(t)
	assert.Nil(err)
}

func TestIsQuarantineActive(t *testing.T) {

	factoryMock := &tmock.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantine := &Quarantine{
		Nodes: []*Node{
			{
				Name:        "foo",
				isolate:     false,
				Daemonsets:  []Daemonset{},
				Deployments: []Deployment{},
				logger:      ctrl.Log.WithName("test"),
				Flags:       &drain.Helper{},
				ioStreams: genericclioptions.IOStreams{
					In:     os.Stdin,
					Out:    os.Stdout,
					ErrOut: os.Stdout,
				},
				Debug: Debug{
					Enabled: false,
				},
			},
		},
		logger: ctrl.Log.WithName("test"),
		Debug: Debug{
			Enabled: false,
		},
		isActive:   false,
		conditions: []metav1.Condition{},
	}

	isActive := quarantine.IsActive()

	assert := assert.New(t)
	assert.True(isActive)
}
