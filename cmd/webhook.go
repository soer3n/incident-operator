package cmd

import (
	"github.com/soer3n/incident-operator/internal/webhook"
	"github.com/spf13/cobra"
)

// NewWebhookCmd represents the api subcommand
func NewWebhookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "webhook",
		Short: "runs backend for webhook",
		Long:  `webhook application`,
		Run: func(cmd *cobra.Command, args []string) {
			webhook.RunWebhookServer()
		},
	}
}
