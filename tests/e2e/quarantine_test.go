package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/soer3n/incident-operator/api/v1alpha1"
)

var quarantineKind *v1alpha1.Quarantine

const quarantineKindName = "quarantine"
const quarantineNodeName = "dev-cluster-worker2"

var _ = Context("Create a quarantine resource", func() {

	Describe("when no existing resource exist", func() {
		It("should start with creating dependencies", func() {
			ctx := context.Background()
			namespace := "test-" + randStringRunes(7)

			By("install a new namespace")
			quarantineNamespace := &v1.Namespace{
				TypeMeta:   metav1.TypeMeta{},
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
