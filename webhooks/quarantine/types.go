package quarantine

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// QuarantineHandler represents struct for validating a quarantine resource
type QuarantineHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
	Log     logr.Logger
}
