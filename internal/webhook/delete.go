package webhook

import (
	"context"
	"errors"
	"log"

	"github.com/soer3n/incident-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteWebhook represents func for deleting resources for a functional webhook
func DeleteWebhook(namespace string) error {

	var err error

	c := utils.GetTypedKubernetesClient()

	log.Print("deleting secrets...")

	delOpts := metav1.DeleteOptions{}

	if err = c.CoreV1().Secrets(namespace).Delete(context.TODO(), "incident-webhook", delOpts); err != nil {
		log.Print(err.Error())
	}

	log.Print("deleting validating admission webhook...")

	if err = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete(context.TODO(), "quarantine", delOpts); err != nil {
		log.Print(err.Error())
	}

	log.Print("deleting mutating admission webhook...")

	if err = c.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(context.TODO(), "quarantine", delOpts); err != nil {
		log.Print(err.Error())
	}

	if err != nil {
		return errors.New("see log for errors during execution")
	}

	log.Print("webhook assets deleted successfully...")
	return nil
}
