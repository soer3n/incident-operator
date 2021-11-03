package cmd

import (
	"github.com/soer3n/incident-operator/cmd/webhook"
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
	cmd.AddCommand(webhook.NewWebhookCreateCertsCmd())
	cmd.AddCommand(webhook.NewWebhookDeleteCertsCmd())
	return cmd
}

func newWebhookServeCmd() *cobra.Command {

	metricsAddr := ":8080"
	probeAddr := ":8081"
	// certDir := "./certs/"
	enableLeaderElection := false

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "runs backend for webhook",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			certDir, err := cmd.Flags().GetString("cert-dir")

			if err != nil {
				return
			}
			webhook.Run(certDir, metricsAddr, probeAddr, enableLeaderElection)
		},
	}

	cmd.PersistentFlags().String("cert-dir", "./certs/", "directory where to find certs")
	return cmd
}
