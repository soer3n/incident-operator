package testcases

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	fakeappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1/fake"
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
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("update", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
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

	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependWatchReactor("pods", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		fakeWatch := watch.NewRaceFreeFake()
		fakeWatch.Action(watch.Added, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: corev1.PodSpec{
				NodeName:   nodeName,
				Containers: []corev1.Container{},
			},
		})
		return true, fakeWatch, nil
	})

	fakeClientset.AppsV1().(*fakeappsv1.FakeAppsV1).PrependReactor("get", "daemonsets", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"key": "value",
					},
				},
			},
		}, nil
	})
	fakeClientset.AppsV1().(*fakeappsv1.FakeAppsV1).PrependReactor("update", "daemonsets", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		}}, nil
	})

	fakeClientset.AppsV1().(*fakeappsv1.FakeAppsV1).PrependReactor("get", "deployments", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"key": "value",
					},
				},
			},
		}, nil
	})
	fakeClientset.AppsV1().(*fakeappsv1.FakeAppsV1).PrependReactor("update", "deployments", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		}}, nil
	})
}
