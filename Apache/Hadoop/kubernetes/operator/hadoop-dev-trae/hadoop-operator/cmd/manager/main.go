package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	hadoopv1alpha1 "hadoop-operator/pkg/apis/hadoop/v1alpha1"
	"hadoop-operator/pkg/controller"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup flags
	leaderElection := flag.Bool("leader-elect", true, "Enable leader election for controller manager")
	flag.Parse()

	// Setup logs
	// zap.New(zap.UseDevMode(true))

	// Add the CRDs to the runtime scheme
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		fmt.Printf("Error adding client-go scheme: %v\n", err)
		os.Exit(1)
	}

	if err := hadoopv1alpha1.AddToScheme(scheme); err != nil {
		fmt.Printf("Error adding hadoop scheme: %v\n", err)
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Error getting config: %v\n", err)
		os.Exit(1)
	}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // Disable metrics
		Port:               9443,
		LeaderElection:     *leaderElection,
		LeaderElectionID:   "hadoop-operator-leader",
	})
	if err != nil {
		fmt.Printf("Error creating manager: %v\n", err)
		os.Exit(1)
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		fmt.Printf("Error adding health check: %v\n", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		fmt.Printf("Error adding ready check: %v\n", err)
		os.Exit(1)
	}

	// Register the controller
	if err := controller.AddToManager(mgr); err != nil {
		fmt.Printf("Error adding controller: %v\n", err)
		os.Exit(1)
	}

	// Start the manager
	fmt.Println("Starting manager...")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		fmt.Printf("Error starting manager: %v\n", err)
		os.Exit(1)
	}
}
