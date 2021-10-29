package webhook

import (
	"context"
	"log"

	"github.com/soer3n/yaho/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteWebhook represents func for deleting resources for a functional webhook
func DeleteWebhook(namespace string) error {

	var err error

	c := client.New().TypedClient

	log.Print("deleting secrets...")

	getOpts := metav1.GetOptions{}
	_, err = c.CoreV1().Secrets(namespace).Get(context.TODO(), "incident-webhook", getOpts)

	if err != nil {
		return err
	}

	delOpts := metav1.DeleteOptions{}
	err = c.CoreV1().Secrets(namespace).Delete(context.TODO(), "incident-webhook", delOpts)

	if err != nil {
		return err
	}

	log.Print("deleting validating admission webhook...")

	_, err = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(context.TODO(), "quarantine", getOpts)

	if err != nil {
		return err
	}

	err = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete(context.TODO(), "quarantine", delOpts)

	if err != nil {
		return err
	}

	log.Print("webhook assets deleted successfully...")

	return nil
}
