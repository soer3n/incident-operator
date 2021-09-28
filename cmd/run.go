package cmd

import (
	"strings"

	"github.com/soer3n/incident-operator/internal/cli"
	"github.com/spf13/cobra"
)

//NewRootCmd represents the root command manager
func NewJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "runs job related to quarantine validation",
		Long:  `jobs`,
	}

	cmd.AddCommand(newJobRescheduleCmd())
	return cmd
}

// NewWebhookCmd represents the api subcommand
func newJobRescheduleCmd() *cobra.Command {
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
