package cli

import (
	"context"
	"encoding/json"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	"github.com/sirupsen/logrus"
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/templates/loader"
	"github.com/soer3n/incident-operator/internal/utils"
)

const quarantineControllerLabelKey = "control-plane"
const quarantineControllerLabelValue = "controller-manager"
const quarantineControllerLabelIgnoreNodeKey = "ops.soer3n.info/isolate"
const quarantineControllerLabelIgnoreNodeValue = "true"

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func New(logger logrus.FieldLogger) *CLI {

	var dr dynamic.ResourceInterface

	config, err := loader.LoadConfig("./config.yaml", logger)

	if err != nil {
		logger.Error(err)
		return nil
	}

	cli := &CLI{
		config: config,
		logger: logger,
		dr:     dr,
		q: &v1alpha1.Quarantine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-q",
				Namespace: "incident-operator-system",
			},
			Spec: v1alpha1.QuarantineSpec{
				Nodes:     []v1alpha1.Node{},
				Resources: []v1alpha1.Resource{},
				Debug: v1alpha1.Debug{
					Enabled: false,
				},
				Flags: v1alpha1.Flags{
					IgnoreAllDaemonSets: pointer.BoolPtr(false),
					DisableEviction:     pointer.BoolPtr(false),
					DeleteEmptyDirData:  pointer.BoolPtr(false),
					Force:               pointer.BoolPtr(false),
					IgnoreErrors:        pointer.BoolPtr(false),
				},
			},
		},
	}

	nodeObjs := cli.getNodes()

	for _, item := range nodeObjs.Items {
		node := v1alpha1.Node{
			Name:    item.Name,
			Isolate: false,
			Flags: v1alpha1.Flags{
				IgnoreAllDaemonSets: pointer.BoolPtr(false),
				DisableEviction:     pointer.BoolPtr(false),
				DeleteEmptyDirData:  pointer.BoolPtr(false),
				Force:               pointer.BoolPtr(false),
				IgnoreErrors:        pointer.BoolPtr(false),
			},
			Resources: []v1alpha1.Resource{},
		}

		cli.q.Spec.Nodes = append(cli.q.Spec.Nodes, node)
	}

	return cli
}

func (cli CLI) GenerateWebhookCerts() error {

	if err := cli.config.SetCerts("", cli.logger); err != nil {
		return err
	}

	return nil
}

func (cli *CLI) InstallResources() error {

	resources, err := loader.LoadManifests(cli.config, cli.logger)

	if err != nil {
		cli.logger.Error(err)
		return err
	}

	for _, r := range resources {

		obj := &unstructured.Unstructured{}

		_, gvk, err := decUnstructured.Decode(r.Raw, nil, obj)

		if err != nil {
			return err
		}

		if err := cli.setDynamicClient(gvk, obj.GetNamespace()); err != nil {
			return err
		}

		cli.logger.Infof("create or update object %s. Kind: %s  APIVersion: %s", obj.GetName(), obj.GetKind(), obj.GetAPIVersion())

		if err := cli.patchResource(r.Raw, obj, cli.logger); err != nil {
			return err
		}
	}

	return nil
}

func (cli *CLI) DeleteResources() error {

	resources, err := loader.LoadManifests(cli.config, cli.logger)

	if err != nil {
		cli.logger.Error(err)
		return err
	}

	for _, r := range resources {

		obj := &unstructured.Unstructured{}

		_, gvk, err := decUnstructured.Decode(r.Raw, nil, obj)

		if err != nil {
			return err
		}

		if err := cli.setDynamicClient(gvk, obj.GetNamespace()); err != nil {
			return err
		}

		cli.logger.Infof("deleting object %s. Kind: %s  APIVersion: %s", obj.GetName(), obj.GetKind(), obj.GetAPIVersion())

		if err := cli.dr.Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (cli CLI) getNodes() *corev1.NodeList {

	var nodes *corev1.NodeList
	var err error

	c := utils.GetTypedKubernetesClient()

	listOpts := metav1.ListOptions{}

	if nodes, err = c.CoreV1().Nodes().List(context.TODO(), listOpts); err != nil {
		cli.logger.Error(err)
		return nil
	}

	return nodes
}

func (cli CLI) getPodsByNode(namespace, node string) *corev1.PodList {

	var pods *corev1.PodList
	var err error

	c := utils.GetTypedKubernetesClient()

	listOpts := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	}

	if pods, err = c.CoreV1().Pods(namespace).List(context.TODO(), listOpts); err != nil {
		cli.logger.Error(err)
		return nil
	}

	return pods
}

func (cli *CLI) setDynamicClient(gvk *schema.GroupVersionKind, namespace string) error {

	client := utils.GetDynamicKubernetesClient()
	dc := utils.GetDiscoveryKubernetesClient()
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)

	if err != nil {
		return err
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		cli.dr = client.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		cli.dr = client.Resource(mapping.Resource)
	}

	return nil

}

func (cli *CLI) patchResource(raw []byte, obj *unstructured.Unstructured, logger logrus.FieldLogger) error {

	data, err := json.Marshal(obj)

	if err != nil {
		return err
	}

	_, err = cli.dr.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "kit",
	})

	if err != nil {
		logger.Fatal(err)
		return err
	}

	return nil
}

func getControllerPod(c kubernetes.Interface) (*corev1.Pod, error) {

	var pods *corev1.PodList
	var pod *corev1.Pod
	var err error

	listOpts := metav1.ListOptions{
		LabelSelector: quarantineControllerLabelKey + "=" + quarantineControllerLabelValue,
	}

	if pods, err = c.CoreV1().Pods("").List(context.TODO(), listOpts); err != nil {
		return pod, err
	}

	if len(pods.Items) > 1 {
		return pod, errors.New("multiple controller pods found")
	}

	return &pods.Items[0], nil
}

func getControllerNode(c kubernetes.Interface, pod *corev1.Pod) (*corev1.Node, error) {

	var node *corev1.Node
	var err error

	getOpts := metav1.GetOptions{}

	if node, err = c.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, getOpts); err != nil {
		return node, err
	}

	return node, nil
}
