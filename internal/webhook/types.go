package webhook

import (
	"k8s.io/api/admission/v1beta1"
)

type QuarantineHandler struct {
	body     []byte
	response *v1beta1.AdmissionReview
}
