/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure they are available for kubeconfig-based authentication.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	hadoopv1 "github.com/hadoop-operator/hadoop-k8s-operator/operator/pkg/apis/hadoop/v1"
	"github.com/hadoop-operator/hadoop-k8s-operator/operator/pkg/controller"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	// Program name
	programName = "hadoop-operator"
	// Version information
	version = "0.1.0"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(hadoopv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
		watchNamespace       string
		devMode              bool
		logLevel             string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&watchNamespace, "watch-namespace", "",
		"Specify a namespace to watch. If empty, all namespaces are watched.")
	flag.BoolVar(&devMode, "dev-mode", false,
		"Enable development mode with reduced reconcile times.")
	flag.StringVar(&logLevel, "log-level", "info",
		"Log level (debug, info, warn, error)")

	opts := zap.Options{
		Development: devMode,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Set log level
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Print version information
	setupLog.Info(fmt.Sprintf("Starting %s", programName))
	setupLog.Info(fmt.Sprintf("Version: %s", version))
	setupLog.Info(fmt.Sprintf("Log level: %s", logLevel))
	setupLog.Info(fmt.Sprintf("Watch namespace: %s", watchNamespace))
	setupLog.Info(fmt.Sprintf("Dev mode: %t", devMode))

	// Determine the namespace to watch
	namespace := watchNamespace
	if namespace == "" {
		setupLog.Info("Watching all namespaces.")
	} else {
		setupLog.Info(fmt.Sprintf("Watching namespace: %s", namespace))
	}

	// Create the manager options
	managerOptions := ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "hadoop-operator-lock",
		LeaderElectionNamespace: namespace,
	}

	// Configure namespace watching if specified
	if namespace != "" {
		managerOptions.Cache = cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
		}
	}

	// Create the manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), managerOptions)
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	// Create the controllers
	hadoopClusterReconciler := controller.NewHadoopClusterReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("hadoop-cluster-controller"),
	)
	hadoopClusterReconciler.LogLevel = logLevel
	hadoopClusterReconciler.DevMode = devMode

	if err = hadoopClusterReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HadoopCluster")
		os.Exit(1)
	}

	hadoopApplicationReconciler := controller.NewHadoopApplicationReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("hadoop-application-controller"),
	)

	if err = hadoopApplicationReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HadoopApplication")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	// Setup webhooks if enabled
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&hadoopv1.HadoopCluster{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "HadoopCluster")
			os.Exit(1)
		}
		setupLog.Info("Webhooks enabled")
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to add healthz check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to add readyz check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
