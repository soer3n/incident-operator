package main

import (
	"flag"
	"log"
	"os"

	opsv1alpha1 "github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/controllers"
	"github.com/soer3n/incident-operator/webhooks/quarantine"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(opsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	command := NewRootCmd()
	if err := command.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

//NewRootCmd represents the root command kit
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "k8s incident operator",
		Long:  `runs kubernetes incident operator`,
	}

	cmd.AddCommand(NewOperatorCmd())
	return cmd
}

//NewOperatorCmd represents the operator subcommand
func NewOperatorCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "operator",
		Short: "runs the operator",
		Long:  `apps operator`,
		Run: func(cmd *cobra.Command, args []string) {

			enableLeaderElection, err := cmd.Flags().GetBool("leader-elect")

			if err != nil {
				return
			}

			metricsAddr, err := cmd.Flags().GetString("metrics-bind-address")

			if err != nil {
				return
			}

			probeAddr, err := cmd.Flags().GetString("health-probe-bind-address")

			if err != nil {
				return
			}

			certDir, err := cmd.Flags().GetString("cert-dir")

			if err != nil {
				return
			}

			runOperator(certDir, metricsAddr, probeAddr, enableLeaderElection)
		},
	}

	cmd.PersistentFlags().String("cert-dir", "/tmp/k8s-webhook-server/serving-certs/", "directory where to find certs")
	cmd.PersistentFlags().String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	cmd.PersistentFlags().String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	cmd.PersistentFlags().Bool("leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	return cmd
}

func runOperator(certDir, metricsAddr, probeAddr string, enableLeaderElection bool) {

	var err error

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "71b71418.soer3n.info",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	var dec *admission.Decoder
	if dec, err = admission.NewDecoder(scheme); err != nil {
		setupLog.Error(err, "unable to setup admission decoder")
		os.Exit(1)
	}

	wh := mgr.GetWebhookServer()
	wh.CertDir = certDir
	wh.Register("/validate", &admission.Webhook{Handler: &quarantine.QuarantineValidateHandler{
		Client:  mgr.GetClient(),
		Decoder: dec,
		Log:     ctrl.Log.WithName("webhook").WithName("ops").WithName("Quarantine"),
	}})
	wh.Register("/mutate", &admission.Webhook{Handler: &quarantine.QuarantineMutateHandler{
		Client:  mgr.GetClient(),
		Decoder: dec,
		Log:     ctrl.Log.WithName("webhook").WithName("ops").WithName("Quarantine"),
	}})

	if err = (&controllers.QuarantineReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ops").WithName("Quarantine"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Quarantine")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
