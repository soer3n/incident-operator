package quarantine

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/watch"
)

func waitForResource(w watch.Interface, logger logr.Logger) {

	defer w.Stop()

	for {
		e := <-w.ResultChan()

		if e.Type == watch.Added || e.Type == watch.Deleted {
			logger.Info("modified...")
		}
	}
}
