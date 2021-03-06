package webhook

import (
	"log"

	"github.com/soer3n/incident-operator/internal/webhook"
	"github.com/spf13/cobra"
)

func NewWebhookDeleteCertsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete resources related to installed webhook",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")

			if err != nil {
				log.Fatal(err.Error())
			}

			if err := webhook.DeleteWebhook(namespace); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	cmd.PersistentFlags().String("namespace", "dev", "namespace for deploying resources")

	return cmd
}

func NewWebhookCreateCertsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "runs job for webhook tls cert creation",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			svc, _ := cmd.Flags().GetString("service")
			namespace, _ := cmd.Flags().GetString("namespace")
			local, _ := cmd.Flags().GetBool("local")
			certDir, _ := cmd.Flags().GetString("certDir")

			if err := webhook.InstallWebhook(svc, namespace, certDir, local); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.PersistentFlags().Bool("local", false, "flag for installing certs on current workstation")
	cmd.PersistentFlags().String("certDir", "/tmp/k8s-webhook-server/serving-certs", "path for local certs")
	cmd.PersistentFlags().String("service", "quarantine-webhook.dev.svc", "name of deployed webhook service")
	cmd.PersistentFlags().String("namespace", "dev", "namespace for deploying resources")

	return cmd
}
