package webhook

import (
	"sync"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes"
)

type QuarantineHandler struct {
	body     []byte
	response *v1beta1.AdmissionReview
	client   kubernetes.Interface
}

type QuarantineHTTPHandler struct {
	mu sync.Mutex
}
