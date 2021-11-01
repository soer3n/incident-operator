package webhook

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

const (
	bitSize    = 4096
	commonName = "Admission Controller Webhook"
)

func (w *Cert) generateWebhookCert() error {

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		panic(err)
	}

	ca := w.getCertTemplate()

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)

	if err != nil {
		return err
	}

	caPEM := new(bytes.Buffer)

	if err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	if err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return err
	}

	w.Ca = CA{
		Key:     caPrivKeyPEM.Bytes(),
		Cert:    caPEM.Bytes(),
		CertObj: ca,
		KeyObj:  key,
	}

	return nil
}

func (w *Cert) getCertTemplate() *x509.Certificate {
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

func (w *Cert) create(CommonName string) error {

	localIP := net.ParseIP("127.0.0.1")

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		IPAddresses:  []net.IP{localIP},
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
	if err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	if err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return err
	}

	w.Cert = certPEM.Bytes()
	w.Key = certPrivKeyPEM.Bytes()

	_, err = tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())

	if err != nil {
		return err
	}

	return nil
}
