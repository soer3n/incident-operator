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
	return cmd
}

func newWebhookServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "runs backend for webhook",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			webhook.RunWebhookServer()
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

func newWebhookCreateCertsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "certs",
		Short: "runs job for webhook tls cert creation",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("%v", webhook.InstallWebhookCerts("quarantine-webhook.dev.svc", "dev"))
		},
	}
}
