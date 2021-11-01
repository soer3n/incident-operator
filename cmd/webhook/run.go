package webhook

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	opsv1alpha1 "github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/webhooks/quarantine"
	//+kubebuilder:scaffold:imports
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

// Run represents starting the quarantine operator
func Run() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var certDir string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&certDir, "cert-dir", "./certs/", "The directory for storing webhook related certs.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0",
		Port:                   9443,
		HealthProbeBindAddress: "0",
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "71b71325.soer3n.info",
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
	wh.Register("/validate", &admission.Webhook{Handler: &quarantine.QuarantineHandler{
		Client:  mgr.GetClient(),
		Decoder: dec,
		Log:     ctrl.Log.WithName("webhook").WithName("ops").WithName("Quarantine"),
	}})

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
