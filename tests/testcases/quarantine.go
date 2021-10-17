package testcases

import (
	"os"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	q "github.com/soer3n/incident-operator/internal/quarantine"
	"github.com/soer3n/incident-operator/tests"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
							Resources: []v1alpha1.Resource{
								{
									Type:      "daemonset",
									Name:      "foo",
									Namespace: "foo",
									Keep:      false,
								},
								{
									Type:      "deployment",
									Name:      "baz",
									Namespace: "baz",
									Keep:      false,
								},
							},
						},
						{
							Name:    "worker2",
							Isolate: true,
							Resources: []v1alpha1.Resource{
								{
									Type:      "daemonset",
									Name:      "boo",
									Namespace: "boo",
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
							Enabled: true,
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
					Enabled: true,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func GetQuarantinePrepareStructs() []tests.QuarantineTestCase {

	res := []tests.QuarantineTestCase{}

	fakeClientsetFoo := fake.NewSimpleClientset()
	configureClientset(fakeClientsetFoo, "foo")

	fakeClientsetBar := fake.NewSimpleClientset()
	configureClientset(fakeClientsetBar, "bar")

	res = append(res, tests.QuarantineTestCase{
		ReturnError: nil,
		Input: &q.Quarantine{
			Nodes: []*q.Node{
				{
					Name:    "foo",
					Isolate: false,
					Daemonsets: []q.Daemonset{
						{
							Name:      "foo",
							Namespace: "foo",
							Keep:      true,
						},
					},
					Deployments: []q.Deployment{
						{
							Name:      "foo",
							Namespace: "foo",
							Keep:      true,
						},
					},
					Logger: ctrl.Log.WithName("test"),
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
						Enabled: true,
					},
				},
			},
			Logger: ctrl.Log.WithName("test"),
			Debug: q.Debug{
				Enabled: true,
			},
			Conditions: []metav1.Condition{},
		},
	})

	res = append(res, tests.QuarantineTestCase{
		ReturnError: nil,
		Input: &q.Quarantine{
			Nodes: []*q.Node{
				{
					Name:        "foo",
					Isolate:     true,
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
					Isolate:     true,
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
						Enabled: true,
					},
				},
			},
			Logger: ctrl.Log.WithName("test"),
			Debug: q.Debug{
				Enabled: false,
			},
			Conditions: []metav1.Condition{},
		},
	})

	return res
}

func GetQuarantineStopStructs() []tests.QuarantineTestCase {

	fakeClientset := fake.NewSimpleClientset()
	configureClientset(fakeClientset, "foo")

	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:    "foo",
						Isolate: false,
						Daemonsets: []q.Daemonset{
							{
								Name:      "foo",
								Namespace: "foo",
								Keep:      true,
							},
						},
						Deployments: []q.Deployment{
							{
								Name:      "foo",
								Namespace: "foo",
								Keep:      true,
							},
						},
						Logger: ctrl.Log.WithName("test"),
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
					Enabled: true,
				},
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func GetQuarantineUpdateStructs() []tests.QuarantineTestCase {

	fakeClientset := fake.NewSimpleClientset()
	configureClientset(fakeClientset, "foo")

	return []tests.QuarantineTestCase{
		{
			ReturnError: nil,
			Input: &q.Quarantine{
				Nodes: []*q.Node{
					{
						Name:    "foo",
						Isolate: false,
						Daemonsets: []q.Daemonset{
							{
								Name:      "foo",
								Namespace: "foo",
								Keep:      true,
							},
						},
						Deployments: []q.Deployment{
							{
								Name:      "foo",
								Namespace: "foo",
								Keep:      true,
							},
						},
						Logger: ctrl.Log.WithName("test"),
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
