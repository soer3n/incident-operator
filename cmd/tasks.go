package cmd

import (
	"log"
	"strings"

	"github.com/soer3n/incident-operator/internal/cli"
	"github.com/spf13/cobra"
)

// NewTaskCmd represents the api subcommand
func NewTasksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "runs tasks related to quarantine",
		Long:  `quarantine manager jobs`,
	}

	cmd.AddCommand(NewJobRescheduleCmd())

	return cmd
}

// NewJobRescheduleCmd represents the api subcommand
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

			if err = cli.RescheduleQuarantineController(strings.Split(excludedNodes, ",")); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}
