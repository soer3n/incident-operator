package testcases

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	clienttesting "k8s.io/client-go/testing"
)

func configureClientset(fakeClientset *fake.Clientset, nodeName string) {

	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		}}, nil
	})
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("patch", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		}}, nil
	})
}
