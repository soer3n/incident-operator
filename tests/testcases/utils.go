package testcases

import (
	"context"
	"log"
	"strings"

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
		})

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
	testGlobalSelectors := &TestClientSelectors{
		ListSelectors:  map[string][]string{},
		FieldSelectors: map[string][]string{},
	}

	for _, n := range t.Namespaces {

		for _, v := range n.Pods {

			v.pod = corev1.Pod{
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
				v.pod.ObjectMeta.Labels["ops.soer3n.info/quarantine"] = "true"
			}

			if v.Resource.Taint {
				v.pod.Spec.Tolerations = []corev1.Toleration{
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
					v.pod,
				},
			}

			if len(v.Resource.ListSelector) > 0 {
				for _, s := range v.Resource.ListSelector {
					p.On("List", context.Background(), metav1.ListOptions{
						LabelSelector: s,
					}).Return(podList, nil)
				}
			}

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

			p.On("Get", context.TODO(), v.Resource.Name, metav1.GetOptions{}).Return(v.pod, nil)

			p.On("Update", context.Background(), v.pod, metav1.UpdateOptions{}).Return(v.pod, nil)
			p.On("Update", context.TODO(), v.pod, metav1.UpdateOptions{}).Return(v.pod, nil)

			gracePeriod := int64(0)

			p.On("Delete", context.TODO(), v.Resource.Name, metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
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
						watchChan.Add(&v.pod)
					}()
				})
			}

		}
		corev1Mock.On("Pods", n.Name).Return(p)
		selectors := getSelectorMaps(n.Pods)
		testGlobalSelectors = mergeMaps(testGlobalSelectors, selectors)
		n.parsePodList(p, selectors)
	}

	t.parsePodList(p, testGlobalSelectors)
}

func mergeMaps(global, namespaced *TestClientSelectors) *TestClientSelectors {

	for k, v := range namespaced.ListSelectors {
		if _, ok := global.ListSelectors[k]; !ok {
			global.ListSelectors[k] = v
			continue
		}

		global.ListSelectors[k] = append(global.ListSelectors[k], v...)
	}

	for k, v := range namespaced.FieldSelectors {
		if _, ok := global.FieldSelectors[k]; !ok {
			global.FieldSelectors[k] = v
			continue
		}

		global.FieldSelectors[k] = append(global.FieldSelectors[k], v...)
	}

	return global
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

func getSelectorMaps(pods []*TestClientPod) *TestClientSelectors {

	tcs := &TestClientSelectors{
		ListSelectors:  map[string][]string{},
		FieldSelectors: map[string][]string{},
	}

	for _, p := range pods {
		parseSelectorMap(p.Resource, tcs)
	}

	return tcs
}

func parseSelectorMap(resource TestClientResource, tcs *TestClientSelectors) {

	if len(resource.FieldSelector) > 0 {
		for _, fs := range resource.FieldSelector {
			if len(fs) == 0 {
				continue
			}

			fieldSelectorList := strings.Split(fs, "=")

			if len(fieldSelectorList) > 2 {
				panic(errors.NewBadRequest("more than one operator!"))
			}

			if _, ok := tcs.ListSelectors[fieldSelectorList[0]]; !ok {
				tcs.ListSelectors[fieldSelectorList[0]] = []string{fieldSelectorList[1]}
				continue
			}

			if Contains(tcs.ListSelectors[fieldSelectorList[0]], fieldSelectorList[1]) {
				continue
			}

			tcs.ListSelectors[fieldSelectorList[0]] = append(tcs.ListSelectors[fieldSelectorList[0]], fieldSelectorList[1])
		}

	}

	if len(resource.ListSelector) > 0 {
		for _, ls := range resource.ListSelector {
			if len(ls) == 0 {
				continue
			}

			selectorList := strings.Split(ls, "=")

			if len(selectorList) > 2 {
				panic(errors.NewBadRequest("more than one operator!"))
			}

			if _, ok := tcs.ListSelectors[selectorList[0]]; !ok {
				tcs.ListSelectors[selectorList[0]] = []string{selectorList[1]}
				continue
			}

			if Contains(tcs.ListSelectors[selectorList[0]], selectorList[1]) {
				continue
			}

			tcs.ListSelectors[selectorList[0]] = append(tcs.ListSelectors[selectorList[0]], selectorList[1])
		}
	}
}

func (n *TestClientNamespace) parsePodList(podv1Mock *mocks.PodV1, selectors *TestClientSelectors) error {

	labelMap := map[string][]*corev1.Pod{}
	log.Print(labelMap)
	/*fieldSelectorMap := map[string][]*corev1.Pod{}

	for _, p := range n.Pods {

	}

	for _, v := range n.Pods {

	}

	for k, v := range labelMap {
		podv1Mock.On("List", context.Background(), metav1.ListOptions{
			LabelSelector: s,
			FieldSelector: f,
		}).Return(podList, nil)
	}
	*/
	return nil
}

func (n *TestClientQuarantine) parsePodList(podv1Mock *mocks.PodV1, selectors *TestClientSelectors) error {

	labelMap := map[string][]*corev1.Pod{}
	log.Print(labelMap)

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
		appsv1Mock.On("Daemonsets", n.Name).Return(ds)
	}

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

func prepareClientMock(clientset *mocks.Client) {

	appsv1Mock := &mocks.AppsV1{}
	corev1Mock := &mocks.CoreV1{}
	policyv1beta1Mock := &mocks.PolicyV1Beta1{}
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
	clientset.On("PolicyV1beta1").Return(policyv1beta1Mock)
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

	nodeC := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "baz",
		},
		Spec: corev1.NodeSpec{
			Unschedulable: false,
		},
	}

	patch := []byte{0x7b, 0x22, 0x73, 0x70, 0x65, 0x63, 0x22, 0x3a, 0x7b, 0x22, 0x75, 0x6e, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x3a, 0x74, 0x72, 0x75, 0x65, 0x7d, 0x7d}

	var list []string

	n.On("Get", context.Background(), "foo", metav1.GetOptions{}).Return(nodeA, nil)
	n.On("Get", context.Background(), "bar", metav1.GetOptions{}).Return(nodeB, nil)
	n.On("Get", context.Background(), "baz", metav1.GetOptions{}).Return(nodeC, nil)

	watchChan := watch.NewFake()
	watchChanTwo := watch.NewFake()
	watchChanThree := watch.NewFake()
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
	}).Once()

	n.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=bar",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChanTwo, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChanTwo.Add(nodeB)
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
	}).Once()

	n.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=foo",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChanTwo, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChanTwo.Add(nodeB)
		}()
	})

	n.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=baz",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChanTwo, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChanTwo.Add(nodeC)
		}()
	}).Once()

	n.On("Watch", context.TODO(), metav1.ListOptions{
		LabelSelector:  "kubernetes.io/hostname=baz",
		Watch:          true,
		TimeoutSeconds: &timeout,
	}).Return(
		watchChanThree, nil,
	).Run(func(args mock.Arguments) {
		go func() {
			watchChanThree.Add(nodeC)
		}()
	})

	n.On("Update", context.Background(), nodeA, metav1.UpdateOptions{}).Return(nodeA, nil)
	n.On("Update", context.Background(), nodeB, metav1.UpdateOptions{}).Return(nodeB, nil)
	n.On("Update", context.Background(), nodeC, metav1.UpdateOptions{}).Return(nodeC, nil)

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
					Operator: "Exists",
					Value:    "",
					Effect:   "NoSchedule",
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

	p.On("List", context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=bar",
	}).Return(podListStart, nil)

	p.On("Update", context.Background(), isolatedPod, metav1.UpdateOptions{}).Return(isolatedPod, nil)
	p.On("Update", context.TODO(), isolatedPod, metav1.UpdateOptions{}).Return(isolatedPod, nil)

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

	patchBar = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x5d}
	ds.On("Patch", context.TODO(), "foo", types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(daemonset, nil)

	patchBar = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x61, 0x64, 0x64, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x6b, 0x65, 0x79, 0x22, 0x3a, 0x22, 0x71, 0x75, 0x61, 0x72, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x22, 0x2c, 0x22, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x4e, 0x6f, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x22, 0x7d, 0x5d, 0x7d, 0x5d}
	ds.On("Patch", context.TODO(), "foo", types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(daemonset, nil)

	ds.On("Get", context.TODO(), "foo", metav1.GetOptions{}).Return(daemonset, nil)
	ds.On("Update", context.Background(), daemonset, metav1.UpdateOptions{}).Return(daemonset, nil)

	return ds
}

func getDeploymentMock() *mocks.DeploymentV1 {

	var list []string

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

	patchBar := []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x5d}
	d.On("Patch", context.TODO(), metav1.PatchOptions{}, types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(deployment, nil)

	patchBar = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x61, 0x64, 0x64, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x6f, 0x70, 0x73, 0x2e, 0x73, 0x6f, 0x65, 0x72, 0x33, 0x6e, 0x2e, 0x69, 0x6e, 0x66, 0x6f, 0x2f, 0x71, 0x75, 0x61, 0x72, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x22, 0x3a, 0x22, 0x74, 0x72, 0x75, 0x65, 0x22, 0x7d, 0x7d, 0x5d}
	d.On("Patch", context.TODO(), metav1.PatchOptions{}, types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(deployment, nil)

	patchBar = []byte{0x5b, 0x7b, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x61, 0x64, 0x64, 0x22, 0x2c, 0x22, 0x70, 0x61, 0x74, 0x68, 0x22, 0x3a, 0x22, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x6b, 0x65, 0x79, 0x22, 0x3a, 0x22, 0x71, 0x75, 0x61, 0x72, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x22, 0x2c, 0x22, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2c, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x4e, 0x6f, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x22, 0x7d, 0x5d, 0x7d, 0x5d}
	d.On("Patch", context.TODO(), metav1.PatchOptions{}, types.JSONPatchType, patchBar, metav1.PatchOptions{}, list).Return(deployment, nil)

	return d
}
