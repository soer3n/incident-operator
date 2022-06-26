package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/templates/loader"
	"k8s.io/client-go/dynamic"
)

type CLI struct {
	config *loader.Config
	logger logrus.FieldLogger
	dr     dynamic.ResourceInterface
	q      *v1alpha1.Quarantine
}

type cellEntry struct {
	value string
	key   string
	desc  string
}
