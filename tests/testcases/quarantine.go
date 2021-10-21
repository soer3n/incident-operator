package testcases

import (
	"os"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	q "github.com/soer3n/incident-operator/internal/quarantine"
	"github.com/soer3n/incident-operator/tests"
	mocks "github.com/soer3n/incident-operator/tests/mocks/typed"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

	fakeClientset := &mocks.Client{}
	prepareClientMock(fakeClientset)

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
	clientsetA := &mocks.Client{}
	prepareClientMock(clientsetA)

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
						Client: clientsetA,
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
						Client: clientsetA,
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

	clientsetB := &mocks.Client{}
	prepareClientMock(clientsetB)
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
						Client: clientsetB,
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
						Client: clientsetB,
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

	fakeClientset := &mocks.Client{}
	prepareClientMock(fakeClientset)

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

	fakeClientset := &mocks.Client{}
	prepareClientMock(fakeClientset)

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
