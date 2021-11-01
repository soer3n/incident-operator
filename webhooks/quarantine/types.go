package quarantine

import (
	"github.com/go-logr/logr"
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// QuarantineHandler represents struct for validating a quarantine resource
type QuarantineHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
	Log     logr.Logger
	Object  *v1alpha1.Quarantine
}
