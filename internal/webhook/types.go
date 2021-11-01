package webhook

import (
	"crypto/rsa"
	"crypto/x509"
)

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
