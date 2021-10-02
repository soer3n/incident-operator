package cmd

import (
	"strings"

	"github.com/soer3n/incident-operator/internal/cli"
	"github.com/spf13/cobra"
)

// NewWebhookCmd represents the api subcommand
func NewJobRescheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reschedule",
		Short: "runs job to reschedule quarantine validation",
		Long:  `rescheduling controller`,
		Run: func(cmd *cobra.Command, args []string) {
			excludedNodes, err := cmd.Flags().GetString("excludedNodes")

			if err != nil {
				return
			}

			cli.RescheduleQuarantineController(strings.Split(excludedNodes, ","))
		},
	}

	cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}
