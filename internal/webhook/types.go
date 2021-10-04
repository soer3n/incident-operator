package webhook

import (
	"crypto/rsa"
	"crypto/x509"
	"sync"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes"
)

// QuarantineHandler represents struct for validating a quarantine resource
type QuarantineHandler struct {
	body     []byte
	response *v1beta1.AdmissionReview
	client   kubernetes.Interface
}

// QuarantineHTTPHandler represents struct for handling validation requests separately
type QuarantineHTTPHandler struct {
	mu sync.Mutex
}

type WebhookCert struct {
	Ca   WebhookCA
	Key  []byte
	Cert []byte
}

type WebhookCA struct {
	Key     []byte
	Cert    []byte
	CertObj *x509.Certificate
	KeyObj  *rsa.PrivateKey
}
