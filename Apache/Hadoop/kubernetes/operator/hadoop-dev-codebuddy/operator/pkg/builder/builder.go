/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied.  See the License for the specific
language governing permissions and limitations under the License.
*/

package builder

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	hadoopv1 "github.com/hadoop-operator/hadoop-k8s-operator/operator/pkg/apis/hadoop/v1"
)

// Builder is the interface for resource builders
type Builder interface {
	// Build creates or updates all resources managed by this builder
	Build(ctx context.Context, c client.Client, hc *hadoopv1.HadoopCluster) error
	// Cleanup removes all resources created by this builder
	Cleanup(ctx context.Context, c client.Client, hc *hadoopv1.HadoopCluster) error
	// GetStatus returns the current status of the managed resources
	GetStatus(ctx context.Context, c client.Client, hc *hadoopv1.HadoopCluster) error
}

// computePhase derives a human-readable phase string from replica counts.
// StatefulSet and Deployment do NOT expose a .Status.Phase field (only Pods do),
// so we infer the phase from ready/total replicas.
func computePhase(readyReplicas, replicas int32) string {
	if replicas == 0 {
		return "Pending"
	}
	if readyReplicas == replicas {
		return "Running"
	}
	if readyReplicas > 0 {
		return "Degraded"
	}
	return "Pending"
}

// BuilderFactory creates builders for different components
type BuilderFactory struct{}

// NewBuilderFactory creates a new BuilderFactory
func NewBuilderFactory() *BuilderFactory {
	return &BuilderFactory{}
}

// GetBuilders returns all builders for a HadoopCluster
func (f *BuilderFactory) GetBuilders() []Builder {
	return []Builder{
		&ConfigMapBuilder{},
		&ServiceAccountBuilder{},
		&NameNodeBuilder{},
		&DataNodeBuilder{},
		&JournalNodeBuilder{},
		&ResourceManagerBuilder{},
		&NodeManagerBuilder{},
	}
}

// GetHBaseBuilders returns builders for HBase components
func (f *BuilderFactory) GetHBaseBuilders() []Builder {
	return []Builder{
		&HBaseMasterBuilder{},
		&HBaseRegionServerBuilder{},
	}
}
