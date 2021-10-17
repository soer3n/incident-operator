package tests

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	clienttesting "k8s.io/client-go/testing"

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

	factoryMock := &mocks.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("list", "pods", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.PodList{Items: []corev1.Pod{}}, nil
	})
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantines := testcases.GetQuarantineStartStructs(fakeClientset)

	for _, obj := range quarantines {

		err := obj.Input.Start()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestPrepareQuarantine(t *testing.T) {

	factoryMock := &mocks.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {

		if action.GetSubresource() == "foo" {
			return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			}}, nil
		}

		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		}}, nil
	})
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("patch", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {

		if action.GetSubresource() == "foo" {
			return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			}}, nil
		}

		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		}}, nil
	})
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantines := testcases.GetQuarantinePrepareStructs(fakeClientset)

	for _, obj := range quarantines {

		err := obj.Input.Prepare()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestStopQuarantine(t *testing.T) {

	factoryMock := &mocks.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {

		if action.GetSubresource() == "foo" {
			return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			}}, nil
		}

		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		}}, nil
	})
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("update", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {

		if action.GetSubresource() == "foo" {
			return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			}}, nil
		}

		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		}}, nil
	})
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependWatchReactor("nodes", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		fakeWatch := watch.NewRaceFreeFake()
		fakeWatch.Action(watch.Added, &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: corev1.NodeSpec{
				Unschedulable: true,
			},
		})
		return true, fakeWatch, nil
	})
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantines := testcases.GetQuarantineStopStructs(fakeClientset)

	for _, obj := range quarantines {

		err := obj.Input.Stop()
		assert := assert.New(t)
		assert.Nil(err)
	}
}

func TestUpdateQuarantine(t *testing.T) {

	factoryMock := &mocks.K8SFactoryMock{}
	fakeClientset := fake.NewSimpleClientset()
	factoryMock.On("KubernetesClientSet").Return(fakeClientset)

	quarantines := testcases.GetQuarantineStopStructs(fakeClientset)

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
