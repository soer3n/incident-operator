package webhook

import (
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
