package tests

import (
	"github.com/soer3n/incident-operator/internal/quarantine"
)

// QuarantineTestCase represents a struct with setup structs, expected return structs and errors of tested funcs
type QuarantineTestCase struct {
	ReturnValue *quarantine.Quarantine
	ReturnError error
	Input       *quarantine.Quarantine
}
