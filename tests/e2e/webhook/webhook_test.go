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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/soer3n/incident-operator/api/v1alpha1"
)

var quarantineKind *v1alpha1.Quarantine

const quarantineKindName = "quarantine"
const quarantineNodeName = "dev-cluster-worker2"

var _ = Context("Create a quarantine resource", func() {

	Describe("when no existing resource exist", func() {
		It("should start with creating dependencies", func() {
			ctx := context.Background()
			namespace := "test-" + randStringRunes(7)

			scheme := runtime.NewScheme()

			err = v1alpha1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			err = v1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			m, err := manager.New(testEnv.Config, manager.Options{
				Scheme:  scheme,
				Port:    33633,
				Host:    testEnv.WebhookInstallOptions.LocalServingHost,
				CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
			})

			Expect(err).NotTo(HaveOccurred())

			err = (&v1alpha1.Quarantine{}).SetupWebhookWithManager(m)
			Expect(err).NotTo(HaveOccurred())

			server := m.GetWebhookServer()
			server.Port = 33633

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				_ = server.Start(ctx)
			}()

			waitForWebhooks()

			By("install a new namespace")
			quarantineNamespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: namespace},
			}

			err = testClient.Create(ctx, quarantineNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed to create test quarantine resource")

			By("creating a new quarantine resource with the specified name and specified url")
			quarantineKind = &v1alpha1.Quarantine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      quarantineKindName,
					Namespace: namespace,
				},
				Spec: v1alpha1.QuarantineSpec{
					Nodes: []v1alpha1.Node{
						{
							Name:    quarantineNodeName,
							Rescale: false,
						},
					},
					Resources: []v1alpha1.Resource{},
					Debug: v1alpha1.Debug{
						Enabled: false,
					},
				},
			}

			err = testClient.Create(context.Background(), quarantineKind)
			Expect(err).NotTo(HaveOccurred(), "failed to create quarantine resource")

			deployment := &v1alpha1.Quarantine{}

			Eventually(
				GetResourceFunc(context.Background(), client.ObjectKey{Name: quarantineKindName, Namespace: namespace}, deployment),
				time.Second*20, time.Millisecond*1500).Should(BeNil())

			By("should remove this quarantine resource with the specified name")

			err = testClient.Delete(context.Background(), quarantineKind)
			Expect(err).NotTo(HaveOccurred(), "failed to delete quarantine resource")

			By("by deletion of namespace test should finish successfully")

			err = testClient.Delete(context.Background(), quarantineNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed to delete namespace for testing")

			Eventually(
				GetResourceFunc(context.Background(), client.ObjectKey{Name: quarantineKindName, Namespace: namespace}, deployment),
				time.Second*20, time.Millisecond*1500).ShouldNot(BeNil())

			cancel()
		})
	})
})

func GetResourceFunc(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Quarantine) func() error {
	return func() error {
		if err := testClient.Get(ctx, key, obj); err != nil {
			return err
		}

		if len(obj.Status.Conditions) > 0 {
			return nil
		}

		return &errors.StatusError{}
	}
}
