package quarantine

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// QuarantineValidateHandler represents struct for validating a quarantine resource
type QuarantineValidateHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
	Log     logr.Logger
}

// QuarantineMutateHandler represents struct for validating a quarantine resource
type QuarantineMutateHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
	Log     logr.Logger
}
