package main

import (
	appcmd "github.com/soer3n/incident-operator/cmd"
	"github.com/spf13/cobra"
)

func main() {
	command := NewRootCmd()
	command.Execute()
}

//NewRootCmd represents the root command manager
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "manager app",
		Long:  `manager app`,
	}

	cmd.AddCommand(appcmd.NewOperatorCmd())
	cmd.AddCommand(appcmd.NewWebhookCmd())
	return cmd
}
