package testcases

import (
	"os"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	q "github.com/soer3n/incident-operator/internal/quarantine"
	"github.com/soer3n/incident-operator/tests"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/kubectl/pkg/drain"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GetQuarantineInitSpec() []tests.QuarantineInitTestCase {
	return []tests.QuarantineInitTestCase{
		{
			ReturnError: nil,
			Input: &v1alpha1.Quarantine{
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
					Resources: []v1alpha1.Resource{
						{
							Type:      "daemonset",
							Name:      "foo",
							Namespace: "foo",
							Keep:      false,
						},
						{
							Type:      "deployment",
							Name:      "bar",
							Namespace: "bar",
							Keep:      false,
						},
					},
				},
			},
		},
		{
			ReturnError: nil,
			Input: &v1alpha1.Quarantine{
				Status: v1alpha1.QuarantineStatus{
					Conditions: []metav1.Condition{
						{
							Type:    "foo",
							Reason:  "foo",
							Message: "foo",
						},
					},
				},
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
					Resources: []v1alpha1.Resource{
						{
							Type:      "daemonset",
							Name:      "foo",
							Namespace: "foo",
							Keep:      false,
						},
						{
							Type:      "deployment",
							Name:      "bar",
							Namespace: "bar",
							Keep:      false,
						},
					},
				},
			},
		},
	}
}

func GetQuarantineStartStructs() []tests.QuarantineTestCase {

	fakeClientset := fake.NewSimpleClientset()
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("list", "pods", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.PodList{Items: []corev1.Pod{}}, nil
	})

	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:        "foo",
						Isolate:     false,
						Daemonsets:  []q.Daemonset{},
						Deployments: []q.Deployment{},
						Logger:      ctrl.Log.WithName("test"),
						Flags: &drain.Helper{
							Client: fakeClientset,
						},
						IOStreams: genericclioptions.IOStreams{
							In:     os.Stdin,
							Out:    os.Stdout,
							ErrOut: os.Stdout,
						},
						Debug: q.Debug{
							Enabled: false,
						},
					},
				},
				Logger: ctrl.Log.WithName("test"),
				Debug: q.Debug{
					Enabled: false,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func GetQuarantinePrepareStructs() []tests.QuarantineTestCase {

	fakeClientsetFoo := fake.NewSimpleClientset()
	fakeClientsetFoo.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		}}, nil
	})
	fakeClientsetFoo.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("patch", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		}}, nil
	})

	fakeClientsetBar := fake.NewSimpleClientset()
	fakeClientsetBar.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		}}, nil
	})
	fakeClientsetBar.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("patch", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		}}, nil
	})

	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:        "foo",
						Isolate:     false,
						Daemonsets:  []q.Daemonset{},
						Deployments: []q.Deployment{},
						Logger:      ctrl.Log.WithName("test"),
						Flags: &drain.Helper{
							Client: fakeClientsetFoo,
						},
						IOStreams: genericclioptions.IOStreams{
							In:     os.Stdin,
							Out:    os.Stdout,
							ErrOut: os.Stdout,
						},
						Debug: q.Debug{
							Enabled: false,
						},
					},
					{
						Name:        "bar",
						Isolate:     false,
						Daemonsets:  []q.Daemonset{},
						Deployments: []q.Deployment{},
						Logger:      ctrl.Log.WithName("test"),
						Flags: &drain.Helper{
							Client: fakeClientsetBar,
						},
						IOStreams: genericclioptions.IOStreams{
							In:     os.Stdin,
							Out:    os.Stdout,
							ErrOut: os.Stdout,
						},
						Debug: q.Debug{
							Enabled: false,
						},
					},
				},
				Logger: ctrl.Log.WithName("test"),
				Debug: q.Debug{
					Enabled: false,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func GetQuarantineStopStructs() []tests.QuarantineTestCase {

	fakeClientset := fake.NewSimpleClientset()
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		}}, nil
	})
	fakeClientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("update", "nodes", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
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

	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:        "foo",
						Isolate:     false,
						Daemonsets:  []q.Daemonset{},
						Deployments: []q.Deployment{},
						Logger:      ctrl.Log.WithName("test"),
						Flags: &drain.Helper{
							Client: fakeClientset,
						},
						IOStreams: genericclioptions.IOStreams{
							In:     os.Stdin,
							Out:    os.Stdout,
							ErrOut: os.Stdout,
						},
						Debug: q.Debug{
							Enabled: false,
						},
					},
				},
				Logger: ctrl.Log.WithName("test"),
				Debug: q.Debug{
					Enabled: false,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func GetQuarantineUpdateStructs() []tests.QuarantineTestCase {
	fakeClientset := fake.NewSimpleClientset()

	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:        "foo",
						Isolate:     false,
						Daemonsets:  []q.Daemonset{},
						Deployments: []q.Deployment{},
						Logger:      ctrl.Log.WithName("test"),
						Flags: &drain.Helper{
							Client: fakeClientset,
						},
						IOStreams: genericclioptions.IOStreams{
							In:     os.Stdin,
							Out:    os.Stdout,
							ErrOut: os.Stdout,
						},
						Debug: q.Debug{
							Enabled: false,
						},
					},
				},
				Logger: ctrl.Log.WithName("test"),
				Debug: q.Debug{
					Enabled: false,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func GetQuarantineIsActiveStructs() []tests.QuarantineTestCase {
	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:        "foo",
						Isolate:     false,
						Daemonsets:  []q.Daemonset{},
						Deployments: []q.Deployment{},
						Logger:      ctrl.Log.WithName("test"),
						Flags:       &drain.Helper{},
						IOStreams: genericclioptions.IOStreams{
							In:     os.Stdin,
							Out:    os.Stdout,
							ErrOut: os.Stdout,
						},
						Debug: q.Debug{
							Enabled: false,
						},
					},
				},
				Logger: ctrl.Log.WithName("test"),
				Debug: q.Debug{
					Enabled: false,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}
