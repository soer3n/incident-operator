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

// Cert represents struct for installing webhook resources after cert creation
type Cert struct {
	Ca   CA
	Key  []byte
	Cert []byte
}

// CA represents struct for needed items for signing certs with a ca
type CA struct {
	Key     []byte
	Cert    []byte
	CertObj *x509.Certificate
	KeyObj  *rsa.PrivateKey
}
