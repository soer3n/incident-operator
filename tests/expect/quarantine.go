package expect

import (
	"os"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	q "github.com/soer3n/incident-operator/internal/quarantine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubectl/pkg/drain"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GetQuarantineInitSpec() []*v1alpha1.Quarantine {
	return []*v1alpha1.Quarantine{
		{
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
		},
	}
}

func GetQuaratineStartStructs(fakeClientset *fake.Clientset) []*q.Quarantine {
	return []*q.Quarantine{
		{
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
	}
}
