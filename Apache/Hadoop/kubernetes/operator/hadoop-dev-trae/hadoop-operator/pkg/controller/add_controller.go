package controller

import (
	"hadoop-operator/pkg/controller/hadoop"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManager adds all controllers to the manager
func AddToManager(mgr manager.Manager) error {
	if err := hadoop.Add(mgr); err != nil {
		return err
	}
	return nil
}
