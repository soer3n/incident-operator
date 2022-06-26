package tests

import (
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/quarantine"
)

// QuarantineTestCase represents a struct with setup structs, expected return structs and errors of tested funcs
type QuarantineTestCase struct {
	ReturnValue *quarantine.Quarantine
	ReturnError error
	Input       *quarantine.Quarantine
}

// QuarantineTestCase represents a struct with setup structs, expected return structs and errors of tested funcs
type QuarantineInitTestCase struct {
	ReturnValue *v1alpha1.Quarantine
	ReturnError error
	Input       *v1alpha1.Quarantine
}
