package webhook

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"

	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/soer3n/yaho/pkg/client"
)

const (
	bitSize    = 4096
	commonName = "Admission Controller Webhook"
)

func generateWebhookCert() (*Cert, error) {

	webhookCert := &Cert{}

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		panic(err)
	}

	ca := getCertTemplate()

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)

	if err != nil {
		return webhookCert, err
	}

	caPEM := new(bytes.Buffer)

	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	webhookCert.Ca = CA{
		Key:     caPrivKeyPEM.Bytes(),
		Cert:    caPEM.Bytes(),
		CertObj: ca,
		KeyObj:  key,
	}

	return webhookCert, nil
}

func getCertTemplate() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

// InstallWebhook represents func for installing needed resources for a functional webhook
func InstallWebhook(subj, namespace string) error {

	log.Print("generating ca bundle...")
	wc, _ := generateWebhookCert()

	log.Print("generating webhook server cert...")
	_ = wc.create(subj)

	typedClient := client.New().TypedClient

	log.Print("create cert secret...")
	if err := wc.deploySecret(namespace, typedClient); err != nil {
		return err
	}

	log.Print("create webhook config...")
	if err := wc.deployValidationWebhook(namespace, typedClient); err != nil {
		return err
	}

	log.Print("webhook assets installed successfully...")
	return nil
}

func (w *Cert) create(CommonName string) error {

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames:     []string{CommonName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)

	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, w.Ca.CertObj, &certPrivKey.PublicKey, w.Ca.KeyObj)

	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	w.Cert = certPEM.Bytes()
	w.Key = certPrivKeyPEM.Bytes()

	_, err = tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())

	if err != nil {
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
