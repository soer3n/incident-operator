/*
Copyright 2021.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package webhook

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	mr "math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/soer3n/incident-operator/webhooks/quarantine"

	opsv1alpha1 "github.com/soer3n/incident-operator/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient, testClient client.Client

var err error
var cfg *rest.Config
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

const quarantineWebhookPort = 33633
const quarantineWebhookPath = "/validate"

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	// Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_KUBE_APISERVER", "../../../testbin/bin/kube-apiserver")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_ETCD", "../../../testbin/bin/etcd")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_KUBECTL", "../../../testbin/bin/kubectl")).To(Succeed())

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	initWebhookConfig()
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	err = opsv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	testClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	close(done)

}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func init() {
	mr.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		n, _ := rand.Int(rand.Reader, (big.NewInt(30)))
		b[i] = letterRunes[n.Uint64()]
	}
	return string(b)
}

func initWebhookConfig() {
	failedTypeV1 := admissionregv1.Fail
	path := "https://127.0.0.1:" + fmt.Sprint(quarantineWebhookPort) + quarantineWebhookPath
	webhookObj := &admissionregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ValidatingWebhookConfiguration",
			APIVersion: "admissionregistration.k8s.io/v1beta1",
		},
		Webhooks: []admissionregv1.ValidatingWebhook{
			{
				Name: "webhook.test.svc",
				ClientConfig: admissionregv1.WebhookClientConfig{
					URL: &path,
				},
				FailurePolicy: &failedTypeV1,
				Rules: []admissionregv1.RuleWithOperations{
					{
						Operations: []admissionregv1.OperationType{"CREATE"},
						Rule: admissionregv1.Rule{
							APIGroups:   []string{"ops.soer3n.info"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"quarantines"},
						},
					},
				},
			},
		},
	}
	testEnv.WebhookInstallOptions = envtest.WebhookInstallOptions{
		ValidatingWebhooks: []client.Object{
			webhookObj,
		},
	}
}

func waitForWebhooks() {
	timeout := 1 * time.Second

	for {
		time.Sleep(1 * time.Second)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(quarantineWebhookPort)), timeout)
		if err != nil {
			continue
		}
		conn.Close()
		return
	}
}

func startWebhookServer() context.CancelFunc {
	scheme := runtime.NewScheme()

	err = opsv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	m, err := manager.New(testEnv.Config, manager.Options{
		Scheme:  scheme,
		Port:    quarantineWebhookPort,
		Host:    testEnv.WebhookInstallOptions.LocalServingHost,
		CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
	})

	Expect(err).NotTo(HaveOccurred())

	server := m.GetWebhookServer()
	dec, _ := admission.NewDecoder(scheme)

	q := &quarantine.QuarantineHandler{
		Client:  getFakeClient(),
		Decoder: dec,
		Log:     logf.Log,
	}

	server.Register("/validate", &admission.Webhook{
		Handler: q,
	})

	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		_ = m.Start(ctx)
	}()

	waitForWebhooks()

	return cancel
}

func getFakeClient() client.WithWatch {

	quarantineControllerList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "foo",
					Labels: map[string]string{
						"component": "incident-controller-manager",
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "dev-cluster-worker",
				},
			},
		},
	}

	objs := []runtime.Object{quarantineControllerList}
	return fake.NewFakeClient(objs...)
}
