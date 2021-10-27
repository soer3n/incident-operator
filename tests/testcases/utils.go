package testcases

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	mocks "github.com/soer3n/incident-operator/tests/mocks/typed"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func prepareClientMock(clientset *mocks.Client) {

	appsv1Mock := &mocks.AppsV1{}
	corev1Mock := &mocks.CoreV1{}
	discoveryMock := &mocks.Discovery{}

	deployv1Mock := getDeploymentMock()
	daemonv1Mock := getDaemonsetMock()
	nodev1Mock := getNodeMock()
	podv1Mock := getPodMock()

	corev1Mock.On("Nodes").Return(nodev1Mock)
	corev1Mock.On("Pods", "").Return(podv1Mock)
	corev1Mock.On("Pods", "foo").Return(podv1Mock)

	appsv1Mock.On("DaemonSets", "foo").Return(daemonv1Mock)
	appsv1Mock.On("Deployments", "").Return(deployv1Mock)
	appsv1Mock.On("Deployments", "foo").Return(deployv1Mock)

	discoveryMock.On("ServerGroups").Return(&metav1.APIGroupList{
		Groups: []metav1.APIGroup{
			{
				Versions: []metav1.GroupVersionForDiscovery{
					{
						GroupVersion: "deployments/v1",
						Version:      "v1",
					},
					{
						GroupVersion: "pods/v1",
						Version:      "v1",
					},
					{
						GroupVersion: "nodes/v1",
						Version:      "v1",
					},
				},
			},
		},
	}, nil)

	clientset.On("CoreV1").Return(corev1Mock)
	clientset.On("AppsV1").Return(appsv1Mock)
	clientset.On("Discovery").Return(discoveryMock)
}

func getNodeMock() *mocks.NodeV1 {
	n := &mocks.NodeV1{}
	nodeA := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: corev1.NodeSpec{
			Unschedulable: true,
		},
	}

	nodeB := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
		Spec: corev1.NodeSpec{
			Unschedulable: false,
		},
	}

	patch := []byte{0x7b, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x75, 0x6e, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x3a, 0x74, 0x72, 0x75, 0x65, 0x7d, 0x7d}

	var list []string

	n.On("Get", context.Background(), "foo", metav1.GetOptions{}).Return(nodeA, nil)
	n.On("Get", context.Background(), "bar", metav1.GetOptions{}).Return(nodeB, nil)

	watchChan := watch.NewFake()
	timeout := int64(20)

	n.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=bar",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChan, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChan.Add(nodeB)
		}()
	})

	n.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=foo",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChan, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChan.Add(nodeB)
		}()
	})

	n.On("Update", context.Background(), nodeA, metav1.UpdateOptions{}).Return(nodeA, nil)
	n.On("Update", context.Background(), nodeB, metav1.UpdateOptions{}).Return(nodeB, nil)

	n.On("Patch", nil, "foo", types.StrategicMergePatchType, patch, metav1.PatchOptions{}, list).Return(nodeA, nil)
	n.On("Patch", nil, "bar", types.StrategicMergePatchType, patch, metav1.PatchOptions{}, list).Return(nodeA, nil)

	patchBar := []byte{0x7b, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x75, 0x6e, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x7d, 0x7d}
	n.On("Patch", nil, "foo", types.StrategicMergePatchType, patchBar, metav1.PatchOptions{}, list).Return(nodeA, nil)

	return n
}

func getPodMock() *mocks.PodV1 {
	p := &mocks.PodV1{}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "quarantine-debug",
			Namespace: "foo",
		},
		Spec: corev1.PodSpec{
			NodeName:   "foo",
			Containers: []corev1.Container{},
		},
	}

	isolatedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "",
			Labels: map[string]string{
				"ops.soer3n.info/quarantine": "true",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:   "foo",
			Containers: []corev1.Container{},
			Tolerations: []corev1.Toleration{
				{
					Key:      "quarantine",
					Operator: "",
					Value:    "true",
					Effect:   "NoExecute",
				},
			},
		},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: corev1.PodSpec{
					NodeName:   "foo",
					Containers: []corev1.Container{},
				},
			},
		},
	}

	podListStart := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
				Spec: corev1.PodSpec{
					NodeName:   "bar",
					Containers: []corev1.Container{},
				},
			},
		},
	}

	p.On("Get", context.TODO(), "foo", metav1.GetOptions{}).Return(pod, nil)
	p.On("Get", context.TODO(), "quarantine-debug", metav1.GetOptions{}).Return(pod, nil)
	p.On("Get", context.TODO(), "quarantine-debug-foo", metav1.GetOptions{}).Return(pod, nil)

	p.On("Get", context.TODO(), "bar", metav1.GetOptions{}).Return(pod, nil).Once()
	p.On("Get", context.TODO(), "bar", metav1.GetOptions{}).Return(pod, errors.NewNotFound(schema.GroupResource{
		Group:    "",
		Resource: "pod",
	}, "bar"))
	p.On("Get", context.TODO(), "quarantine-debug-bar", metav1.GetOptions{}).Return(pod, errors.NewNotFound(schema.GroupResource{
		Group:    "",
		Resource: "pod",
	}, "quarantine-debug-bar"))

	p.On("List", context.Background(), metav1.ListOptions{
		LabelSelector: "ops.soer3n.info/key=value",
	}).Return(podList, nil)
	p.On("List", context.Background(), metav1.ListOptions{
		LabelSelector: "ops.soer3n.info/quarantine=true",
	}).Return(podList, nil)
	p.On("List", context.Background(), metav1.ListOptions{
		LabelSelector: "key=value",
	}).Return(podList, nil)
	p.On("List", context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=foo",
	}).Return(podList, nil)
	p.On("List", context.Background(), metav1.ListOptions{
		LabelSelector: "ops.soer3n.info/key=value,kubernetes.io/hostname=foo",
	}).Return(podList, nil)

	p.On("List", context.Background(), metav1.ListOptions{
		LabelSelector: "quarantine-start",
		FieldSelector: "spec.nodeName=bar",
	}).Return(podListStart, nil)

	p.On("Update", context.Background(), isolatedPod, metav1.UpdateOptions{}).Return(isolatedPod, nil)

	gracePeriod := int64(0)

	p.On("Delete", context.TODO(), "quarantine-debug-foo", metav1.DeleteOptions{}).Return(nil)
	p.On("Delete", context.TODO(), "foo", metav1.DeleteOptions{}).Return(nil)
	p.On("Delete", context.TODO(), "foo", metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}).Return(nil)
	p.On("Delete", context.TODO(), "bar", metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}).Return(nil)

	watchChan := watch.NewFake()
	timeout := int64(20)

	p.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=bar",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChan, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChan.Add(pod)
		}()
	})

	return p
}

func getDaemonsetMock() *mocks.DaemonsetV1 {
	ds := &mocks.DaemonsetV1{}
	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"key": "value",
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{
						{
							Value:    "foo",
							Key:      "bar",
							Operator: "Exists",
							Effect:   corev1.TaintEffectNoExecute,
						},
					},
				},
			},
		},
	}

	var list []string

	patchBar := []byte{0x7b, 0x22, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x7b, 0x22, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x7d, 0x2c, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x6b, 0x65, 0x79, 0x22, 0x3a, 0x22, 0x62, 0x61, 0x72, 0x22, 0x2c, 0x22, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x66, 0x6f, 0x6f, 0x22, 0x2c, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x4e, 0x6f, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x22, 0x7d, 0x5d, 0x7d, 0x7d, 0x2c, 0x22, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x22, 0x3a, 0x7b, 0x7d, 0x7d}
	ds.On("Patch", context.TODO(), "foo", types.StrategicMergePatchType, patchBar, metav1.PatchOptions{}, list).Return(daemonset, nil)

	ds.On("Get", context.TODO(), "foo", metav1.GetOptions{}).Return(daemonset, nil)
	ds.On("Update", context.Background(), daemonset, metav1.UpdateOptions{}).Return(daemonset, nil)

	return ds
}

func getDeploymentMock() *mocks.DeploymentV1 {
	d := &mocks.DeploymentV1{}
	deployment := &appsv1.Deployment{
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
	}

	d.On("Get", context.TODO(), "foo", metav1.GetOptions{}).Return(deployment, nil)
	d.On("Update", context.Background(), deployment, metav1.UpdateOptions{}).Return(deployment, nil)

	return d
}
