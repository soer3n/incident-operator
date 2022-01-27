package testcases

import (
	"context"
	"strings"

	"gonum.org/v1/gonum/stat/combin"

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

func newQuarantineClient() *TestClientQuarantine {

	clientset := &mocks.Client{}
	return &TestClientQuarantine{
		FakeClient: clientset,
	}
}

func (t *TestClientQuarantine) prepare() {

	appsv1Mock := &mocks.AppsV1{}
	corev1Mock := &mocks.CoreV1{}
	policyv1beta1Mock := &mocks.PolicyV1Beta1{}
	discoveryMock := &mocks.Discovery{}

	t.setNodes(corev1Mock)
	t.setPods(corev1Mock)
	t.setDeployments(appsv1Mock)
	t.setDaemonsets(appsv1Mock)
	t.setDiscoveryClient()

	t.FakeClient.On("CoreV1").Return(corev1Mock)
	t.FakeClient.On("AppsV1").Return(appsv1Mock)
	t.FakeClient.On("PolicyV1beta1").Return(policyv1beta1Mock)
	t.FakeClient.On("Discovery").Return(discoveryMock)
}

func (t *TestClientQuarantine) setNodes(corev1Mock *mocks.CoreV1) {

	var list []string

	n := &mocks.NodeV1{}

	for _, v := range t.Nodes {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: v.Name,
			},
			Spec: corev1.NodeSpec{
				Unschedulable: v.Drain,
			},
		}

		n.On("Get", context.Background(), v.Name, metav1.GetOptions{}).Return(node, nil)

		watchChan := watch.NewFake()
		watchChanTwo := watch.NewFake()
		watchChanThree := watch.NewFake()

		timeout := int64(20)

		n.On("Watch", context.TODO(), metav1.ListOptions{
			LabelSelector:  "kubernetes.io/hostname=" + v.Name,
			Watch:          true,
			TimeoutSeconds: &timeout,
		}).Return(
			watchChan, nil,
		).Run(func(args mock.Arguments) {
			go func() {
				watchChan.Add(node)
			}()
		}).Once()

		n.On("Watch", context.TODO(), metav1.ListOptions{
			LabelSelector:  "kubernetes.io/hostname=" + v.Name,
			Watch:          true,
			TimeoutSeconds: &timeout,
		}).Return(
			watchChanTwo, nil,
		).Run(func(args mock.Arguments) {
			go func() {
				watchChanTwo.Add(node)
			}()
		}).Once()

		n.On("Watch", context.TODO(), metav1.ListOptions{
			LabelSelector:  "kubernetes.io/hostname=" + v.Name,
			Watch:          true,
			TimeoutSeconds: &timeout,
		}).Return(
			watchChanThree, nil,
		).Run(func(args mock.Arguments) {
			go func() {
				watchChanThree.Add(node)
			}()
		}).Once()

		n.On("Update", context.Background(), node, metav1.UpdateOptions{}).Return(node, nil)

		patch := []byte{0x7b, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x75, 0x6e, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x3a, 0x74, 0x72, 0x75, 0x65, 0x7d, 0x7d}
		n.On("Patch", nil, v.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, list).Return(node, nil)

		patch = []byte{0x7b, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x75, 0x6e, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x7d, 0x7d}
		n.On("Patch", nil, v.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, list).Return(node, nil)
	}

	corev1Mock.On("Nodes").Return(n)
}

func (t *TestClientQuarantine) setPods(corev1Mock *mocks.CoreV1) {

	p := &mocks.PodV1{}

	for _, n := range t.Namespaces {

		for _, v := range n.Pods {

			currentPod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Resource.Name,
					Namespace: n.Name,
				},
				Spec: corev1.PodSpec{
					NodeName:   v.Resource.Node,
					Containers: []corev1.Container{},
				},
			}

			if v.Resource.Isolated {
				currentPod.ObjectMeta.Labels["ops.soer3n.info/quarantine"] = "true"
			}

			if v.Resource.Taint {
				currentPod.Spec.Tolerations = []corev1.Toleration{
					{
						Key:      "quarantine",
						Operator: "Exists",
						Value:    "",
						Effect:   "NoSchedule",
					},
				}
			}

			podList := &corev1.PodList{
				Items: []corev1.Pod{
					currentPod,
				},
			}

			if len(v.Resource.ListSelector) > 0 {
				for i := 0; i <= len(v.Resource.ListSelector); i++ {
					c := combin.Permutations(len(v.Resource.ListSelector), i)

					for _, perm := range c {

						if len(perm) < 1 {
							continue
						}

						currentSelector := ""
						for ik, ix := range perm {
							if ik > 0 && len(perm) > 1 {
								currentSelector = currentSelector + ","
							}
							currentSelector = currentSelector + v.Resource.ListSelector[ix]
						}

						p.On("List", context.Background(), metav1.ListOptions{
							LabelSelector: currentSelector,
						}).Return(podList, nil)
					}
				}
			}

			v.pod = &currentPod

			if len(v.Resource.ListSelector) > 0 && len(v.Resource.FieldSelector) > 0 {
				for _, s := range v.Resource.ListSelector {
					for _, f := range v.Resource.FieldSelector {
						p.On("List", context.Background(), metav1.ListOptions{
							LabelSelector: s,
							FieldSelector: f,
						}).Return(podList, nil)
					}
				}
			}

			if len(v.Resource.FieldSelector) > 0 {
				for _, f := range v.Resource.FieldSelector {
					p.On("List", context.Background(), metav1.ListOptions{
						FieldSelector: f,
					}).Return(podList, nil)
				}
			}

			p.On("Get", context.TODO(), v.Resource.Name, metav1.GetOptions{}).Return(v.pod, nil).Once()
			p.On("Get", context.TODO(), v.Resource.Name, metav1.GetOptions{}).Return(v.pod, errors.NewNotFound(schema.GroupResource{}, v.Resource.Name))

			p.On("Create", context.TODO(), mock.MatchedBy(func(pod *corev1.Pod) bool {
				return true
			}), metav1.CreateOptions{}).Return(v.pod, nil)

			p.On("Update", context.Background(), mock.MatchedBy(func(pod *corev1.Pod) bool {
				return true
			}), metav1.UpdateOptions{}).Return(v.pod, nil)
			p.On("Update", context.TODO(), mock.MatchedBy(func(pod *corev1.Pod) bool {
				return true
			}), metav1.UpdateOptions{}).Return(v.pod, nil)

			gracePeriod := int64(0)
			addressedGracePeriod := &gracePeriod

			if !v.Resource.GracePeriod {
				addressedGracePeriod = nil
			}

			p.On("Delete", context.TODO(), v.Resource.Name, metav1.DeleteOptions{
				GracePeriodSeconds: addressedGracePeriod,
			}).Return(nil)

			if v.Resource.Watch {
				watchChan := watch.NewFake()
				timeout := int64(20)

				p.On("Watch", context.TODO(), metav1.ListOptions{
					LabelSelector:  "kubernetes.io/hostname=" + v.Resource.Node,
					Watch:          true,
					TimeoutSeconds: &timeout,
				}).Return(
					watchChan, nil,
				).Run(func(args mock.Arguments) {
					go func() {
						watchChan.Add(v.pod)
					}()
				})
			}

		}

		selectorMap := n.parsePodList(p)
		n.setPodList(p, selectorMap)
		corev1Mock.On("Pods", n.Name).Return(p)
	}

	corev1Mock.On("Pods", "").Return(p)
}

// Contains represents func for checking if a string is in a list of strings
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func (n *TestClientNamespace) parsePodList(podv1Mock *mocks.PodV1) TestClientSelectorStruct {

	fieldSelectorMap := map[string]TestClientSelectorValues{}
	labelSelectorMap := map[string]TestClientSelectorValues{}

	for _, p := range n.Pods {
		if len(p.pod.ObjectMeta.Labels) > 0 {
			for k, v := range p.pod.ObjectMeta.Labels {
				if copy, ok := labelSelectorMap[k]; !ok {
					labelSelectorMap[k] = TestClientSelectorValues{
						Value: v,
						Pods:  []corev1.Pod{*p.pod.DeepCopy()},
					}
					continue
				} else {

					copy.Pods = append(copy.Pods, *p.pod.DeepCopy())
					labelSelectorMap[k] = copy
				}
			}
		}
	}

	for _, p := range n.Pods {
		if len(p.Resource.FieldSelector) > 0 {
			for _, f := range p.Resource.FieldSelector {
				fsList := strings.Split(f, "=")
				if len(fsList) != 2 {
					continue
				}

				if copy, ok := fieldSelectorMap[fsList[0]]; !ok {
					fieldSelectorMap[fsList[0]] = TestClientSelectorValues{
						Value: fsList[1],
						Pods:  []corev1.Pod{*p.pod.DeepCopy()},
					}
					continue
				} else {

					copy.Pods = append(copy.Pods, *p.pod.DeepCopy())
					fieldSelectorMap[fsList[0]] = copy
				}
			}
		}
	}

	return TestClientSelectorStruct{
		FieldSelectors: fieldSelectorMap,
		ListSelectors:  labelSelectorMap,
	}
}

func (n *TestClientNamespace) setPodList(podv1Mock *mocks.PodV1, labelSelectorMap TestClientSelectorStruct) error {
	for k, v := range labelSelectorMap.ListSelectors {

		podList := corev1.PodList{
			Items: v.Pods,
		}

		podv1Mock.On("List", context.Background(), metav1.ListOptions{
			LabelSelector: k + "=" + v.Value,
		}).Return(podList, nil)

		for x, z := range labelSelectorMap.FieldSelectors {

			podv1Mock.On("List", context.Background(), metav1.ListOptions{
				FieldSelector: x + "=" + z.Value,
				LabelSelector: k + "=" + v.Value,
			}).Return(podList, nil)

			podv1Mock.On("List", context.Background(), metav1.ListOptions{
				FieldSelector: x + "=" + z.Value,
			}).Return(podList, nil)
		}
	}

	return nil
}

func (t *TestClientQuarantine) setDeployments(appsv1Mock *mocks.AppsV1) error {

	var list []string

	d := &mocks.DeploymentV1{}

	for _, n := range t.Namespaces {
		for _, v := range n.Deployments {

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: n.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"key": "value",
						},
					},
				},
			}

			d.On("Get", context.TODO(), v.Name, metav1.GetOptions{}).Return(deployment, nil)
			d.On("Update", context.Background(), deployment, metav1.UpdateOptions{}).Return(deployment, nil)

			patch := []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x5d}
			d.On("Patch", context.TODO(), metav1.PatchOptions{}, types.JSONPatchType, patch, metav1.PatchOptions{}, list).Return(deployment, nil)

			patch = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x61, 0x64, 0x64, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x6f, 0x70, 0x73, 0x2e, 0x73, 0x6f, 0x65, 0x72, 0x33, 0x6e, 0x2e, 0x69, 0x6e, 0x66, 0x6f, 0x2f, 0x71, 0x75, 0x61, 0x72, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x22, 0x3a, 0x22, 0x74, 0x72, 0x75, 0x65, 0x22, 0x7d, 0x7d, 0x5d}
			d.On("Patch", context.TODO(), metav1.PatchOptions{}, types.JSONPatchType, patch, metav1.PatchOptions{}, list).Return(deployment, nil)

			patch = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x61, 0x64, 0x64, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x6b, 0x65, 0x79, 0x22, 0x3a, 0x22, 0x71, 0x75, 0x61, 0x72, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x22, 0x2c, 0x22, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x4e, 0x6f, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x22, 0x7d, 0x5d, 0x7d, 0x5d}
			d.On("Patch", context.TODO(), metav1.PatchOptions{}, types.JSONPatchType, patch, metav1.PatchOptions{}, list).Return(deployment, nil)

		}
		appsv1Mock.On("Deployments", n.Name).Return(d)

	}

	appsv1Mock.On("Deployments", "").Return(d)
	return nil
}

func (t *TestClientQuarantine) setDaemonsets(appsv1Mock *mocks.AppsV1) error {

	ds := &mocks.DaemonsetV1{}

	for _, n := range t.Namespaces {
		for _, v := range n.Daemonsets {
			daemonset := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: n.Name,
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"key": "value",
						},
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{},
					},
				},
			}

			if v.Taint {
				daemonset.Spec.Template.Spec.Tolerations = []corev1.Toleration{
					{
						Value:    "foo",
						Key:      "bar",
						Operator: "Exists",
						Effect:   corev1.TaintEffectNoExecute,
					},
				}
			}

			var list []string

			patchBar := []byte{0x7b, 0x22, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x7b, 0x22, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x7d, 0x2c, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x6b, 0x65, 0x79, 0x22, 0x3a, 0x22, 0x62, 0x61, 0x72, 0x22, 0x2c, 0x22, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x66, 0x6f, 0x6f, 0x22, 0x2c, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x4e, 0x6f, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x22, 0x7d, 0x5d, 0x7d, 0x7d, 0x2c, 0x22, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x22, 0x3a, 0x7b, 0x7d, 0x7d}
			ds.On("Patch", context.TODO(), v.Name, types.StrategicMergePatchType, patchBar, metav1.PatchOptions{}, list).Return(daemonset, nil)

			patchBar = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x5d}
			ds.On("Patch", context.TODO(), v.Name, types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(daemonset, nil)

			patchBar = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x61, 0x64, 0x64, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x6b, 0x65, 0x79, 0x22, 0x3a, 0x22, 0x71, 0x75, 0x61, 0x72, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x22, 0x2c, 0x22, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x4e, 0x6f, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x22, 0x7d, 0x5d, 0x7d, 0x5d}
			ds.On("Patch", context.TODO(), v.Name, types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(daemonset, nil)

			ds.On("Get", context.TODO(), v.Name, metav1.GetOptions{}).Return(daemonset, nil)
			ds.On("Update", context.Background(), daemonset, metav1.UpdateOptions{}).Return(daemonset, nil)
		}
		appsv1Mock.On("DaemonSets", n.Name).Return(ds)
	}

	appsv1Mock.On("DaemonSets", "").Return(ds)
	return nil
}

func (t *TestClientQuarantine) setDiscoveryClient() error {

	discoveryMock := &mocks.Discovery{}

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

	t.FakeClient.On("Discovery").Return(discoveryMock)
	return nil
}
