package main

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/soer3n/incident-operator/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	command := NewRootCmd()
	if err := command.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

//NewRootCmd represents the root command kit
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kit",
		Short: "k8s incident toolset",
		Long:  `kubernetes incident toolset`,
	}

	logger := logrus.StandardLogger()

	cmd.AddCommand(NewInstallCmd(logger))
	cmd.AddCommand(NewPrepareCmd(logger))
	cmd.AddCommand(NewEditCmd(logger))
	cmd.AddCommand(NewUnInstallCmd(logger))
	cmd.AddCommand(NewStartCmd(logger))
	cmd.AddCommand(NewStopCmd(logger))
	return cmd
}

// NewIntallCmd represents the api subcommand
func NewInstallCmd(logger logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "installs resources into cluster",
		Long:  `installs needed kubernetes resources into cluster`,
		Run: func(cmd *cobra.Command, args []string) {

			c := cli.New(logger)

			if err := c.GenerateWebhookCerts(); err != nil {
				logger.Error(err)
				return
			}

			if err := c.InstallResources(); err != nil {
				logger.Error(err)
			}
		},
	}

	// cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}

// NewPrepareCmd represents the api subcommand
func NewPrepareCmd(logger logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "prepares custom resource",
		Long:  `prepares custom resource which are deployed into cluster`,
		Run: func(cmd *cobra.Command, args []string) {

			c := cli.New(logger)

			app, err := c.RenderPrepareView()

			if err != nil {
				fmt.Printf("Error rendering terminal view: %s\n", err)
			}

			if err := app.Run(); err != nil {
				fmt.Printf("Error running application: %s\n", err)
			}
		},
	}

	// cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}

// NewEditCmd represents the api subcommand
func NewEditCmd(logger logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edits custom resource",
		Long:  `edits custom resource which are deployed into cluster`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	// cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}

// NewUnIntallCmd represents the api subcommand
func NewUnInstallCmd(logger logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "remove resources into cluster",
		Long:  `remove managed kubernetes resources into cluster`,
		Run: func(cmd *cobra.Command, args []string) {

			c := cli.New(logger)

			if err := c.DeleteResources(); err != nil {
				logger.Error(err)
			}
		},
	}

	// cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}

// NewStartCmd represents the api subcommand
func NewStartCmd(logger logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "starts configured quarantine",
		Long:  `starts configured quarantine`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}

// NewStopCmd represents the api subcommand
func NewStopCmd(logger logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "reset configured quarantine",
		Long:  `reset configured quarantine`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// cmd.PersistentFlags().String("excludedNodes", "", "list of nodes which should be ignored for rescheduling")

	return cmd
}
