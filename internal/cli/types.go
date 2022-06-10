package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/soer3n/incident-operator/internal/templates/loader"
	"k8s.io/client-go/dynamic"
)

type CLI struct {
	config *loader.Config
	logger logrus.FieldLogger
	dr     dynamic.ResourceInterface
}
