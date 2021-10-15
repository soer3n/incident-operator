/*
Copyright 2021.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"github.com/spf13/cobra"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/soer3n/incident-operator/cmd/operator"
	//+kubebuilder:scaffold:imports
)

//NewOperatorCmd represents the operator subcommand
func NewOperatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "runs the operator",
		Long:  `apps operator`,
	}

	cmd.AddCommand(newOperatorCmdServe())
	return cmd
}

//NewOperatorCmd represents the operator subcommand
func newOperatorCmdServe() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "runs the operator",
		Long:  `apps operator`,
		Run: func(cmd *cobra.Command, args []string) {
			operator.Run()
		},
	}
}
