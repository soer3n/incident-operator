package webhook

import (
	"context"
	"log"

	"github.com/soer3n/yaho/pkg/client"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/soer3n/incident-operator/internal/utils"
)

// InstallWebhook represents func for installing needed resources for a functional webhook
func InstallWebhook(subj, namespace, certDir string, local bool) error {

	log.Print("generating ca bundle...")
	wc := &Cert{}

	if err := wc.generateWebhookCert(); err != nil {
		return err
	}

	log.Print("generating webhook server cert...")
	_ = wc.create(subj)

	if local {
		if err := wc.createLocalWebhookCerts(certDir); err != nil {
			return err
		}
		log.Print("webhook assets installed successfully...")
		return nil
	}

	if err := wc.createClusterWebhookCerts(namespace); err != nil {
		return err
	}

	log.Print("webhook assets installed successfully...")
	return nil
}

func (wc *Cert) createLocalWebhookCerts(certDir string) error {
	log.Print("generating cert files...")

	if err := utils.WriteFile("tls.crt", certDir, wc.Cert); err != nil {
		return err
	}

	if err := utils.WriteFile("tls.key", certDir, wc.Key); err != nil {
		return err
	}

	if err := utils.WriteFile("ca.crt", certDir, wc.Ca.Cert); err != nil {
		return err
	}

	return nil
}

func (wc *Cert) createClusterWebhookCerts(namespace string) error {
	log.Print("generating cert files...")

	typedClient := client.New().TypedClient

	log.Print("create cert secret...")
	if err := wc.deploySecret(namespace, typedClient); err != nil {
		return err
	}

	log.Print("create webhook config...")
	if err := wc.deployValidationWebhook(namespace, typedClient); err != nil {
		return err
	}

	return nil
}

func (w Cert) deploySecret(namespace string, c kubernetes.Interface) error {

	var err error

	getOpts := metav1.GetOptions{}
	_, err = c.CoreV1().Secrets(namespace).Get(context.TODO(), "incident-webhook", getOpts)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "incident-webhook",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"tls.crt": w.Cert,
			"tls.key": w.Key,
		},
	}

	if err != nil {
		if errors.IsNotFound(err) {

			createOpts := metav1.CreateOptions{}

			if _, err = c.CoreV1().Secrets(namespace).Create(context.TODO(), secret, createOpts); err != nil {
				return err
			}
		}
		return err
	}

	updateOpts := metav1.UpdateOptions{}

	if _, err = c.CoreV1().Secrets(namespace).Update(context.TODO(), secret, updateOpts); err != nil {
		return err
	}

	return nil
}

func (w Cert) deployValidationWebhook(namespace string, c kubernetes.Interface) error {

	var err error

	getOpts := metav1.GetOptions{}
	_, err = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(context.TODO(), "quarantine", getOpts)

	f := admissionv1beta1.Fail
	validatePath := "/validate"
	webhookConfig := &admissionv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "quarantine",
		},
		Webhooks: []admissionv1beta1.ValidatingWebhook{
			{
				Name: "quarantine.webhook.svc",
				AdmissionReviewVersions: []string{
					"v1",
					"v1beta1",
				},
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      "quarantine-webhook",
						Namespace: namespace,
						Path:      &validatePath,
					},
					CABundle: w.Ca.Cert,
				},
				FailurePolicy: &f,
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							"CREATE",
						},
						Rule: admissionv1beta1.Rule{
							APIGroups: []string{
								"ops.soer3n.info",
							},
							APIVersions: []string{
								"v1alpha1",
							},
							Resources: []string{
								"quarantines",
							},
						},
					},
				},
			},
		},
	}

	if err != nil {
		if errors.IsNotFound(err) {

			createOpts := metav1.CreateOptions{}

			if _, err = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(context.TODO(), webhookConfig, createOpts); err != nil {
				return err
			}
		}
		return err
	}

	updateOpts := metav1.UpdateOptions{}

	if _, err = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(context.TODO(), webhookConfig, updateOpts); err != nil {
		return err
	}

	return nil
}
