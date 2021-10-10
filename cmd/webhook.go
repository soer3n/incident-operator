package cmd

import (
	"log"

	"github.com/soer3n/incident-operator/internal/webhook"
	"github.com/spf13/cobra"
)

// NewWebhookCmd represents the api subcommand
func NewWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "webhook related commands",
		Long:  `webhook application`,
	}

	cmd.AddCommand(newWebhookServeCmd())
	cmd.AddCommand(newWebhookCreateCertsCmd())
	cmd.AddCommand(newWebhookInstallCertsCmd())
	cmd.AddCommand(newWebhookDeleteCertsCmd())
	return cmd
}

func newWebhookServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "runs backend for webhook",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := webhook.RunWebhookServer(); err != nil {
				log.Fatal(err.Error())
			}
		},
	}
}

func newWebhookInstallCertsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "runs job for webhook server creation",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}

func newWebhookDeleteCertsCmd() *cobra.Command {
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

func newWebhookCreateCertsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "runs job for webhook tls cert creation",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			svc, _ := cmd.Flags().GetString("service")
			namespace, _ := cmd.Flags().GetString("namespace")
			log.Printf("%v", webhook.InstallWebhook(svc, namespace))
		},
	}

	cmd.PersistentFlags().String("service", "quarantine-webhook.dev.svc", "name of deployed webhook service")
	cmd.PersistentFlags().String("namespace", "dev", "namespace for deploying resources")

	return cmd
}
