// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
	"github.com/go-logr/logr"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources/builder"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// GVisorInstallationSecretName is the name of the secret ofr the installation.
	GVisorInstallationSecretName = "extension-runtime-gvisor-installation"
	// GVisorSecretName is the name of the secret.
	GVisorSecretName = "extension-runtime-gvisor"
	// GVisorInstallationManagedResourceName is the name of the managed resource installation.
	GVisorInstallationManagedResourceName = "extension-runtime-gvisor-installation"
	// GVisorManagedResourceName is the name of the managed resource.
	GVisorManagedResourceName = "extension-runtime-gvisor"
)

// Reconcile implements ContainerRuntime.Actuator.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("could not create chart renderer for shoot '%s', %w", cr.Namespace, err)
	}

	log.Info("Preparing gVisor installation", "shoot", cluster.Shoot.Name, "shootNamespace", cluster.Shoot.Namespace)
	// create MR containing the prerequisites for the installation DaemonSet
	pspDisabled := gardencorev1beta1helper.IsPSPDisabled(cluster.Shoot)
	gVisorChart, err := charts.RenderGVisorChart(chartRenderer, cluster.Shoot.Spec.Kubernetes.Version, pspDisabled)
	if err != nil {
		return err
	}

	gvisorSecret, gvisorSecretRefs := buildSecret(a.client, gVisorChart, cr.Namespace, GVisorSecretName)
	if err := gvisorSecret.Reconcile(ctx); err != nil {
		return err
	}

	if err := builder.
		NewManagedResource(a.client).
		WithNamespacedName(cr.Namespace, GVisorManagedResourceName).
		WithSecretRefs(gvisorSecretRefs).
		Reconcile(ctx); err != nil {
		return fmt.Errorf("failed to create managed resource - prerequisite for the installation of gVisor, %w", err)
	}

	log.Info("Installing gVisor", "shoot", cluster.Shoot.Name, "shootNamespace", cluster.Shoot.Namespace, "workerPoolName", cr.Spec.WorkerPool.Name)
	gVisorInstallationChart, err := charts.RenderGVisorInstallationChart(chartRenderer, cr)
	if err != nil {
		return err
	}

	gVisorInstallationSecret, gVisorInstallationSecretRefs := buildSecret(a.client, gVisorInstallationChart, cr.Namespace, fmt.Sprintf("%s-%s", GVisorInstallationSecretName, cr.Spec.WorkerPool.Name))
	if err := gVisorInstallationSecret.Reconcile(ctx); err != nil {
		return err
	}

	return builder.
		NewManagedResource(a.client).
		WithNamespacedName(cr.Namespace, GetGVisorInstallationManagedResourceName(cr)).
		WithSecretRefs(gVisorInstallationSecretRefs).
		Reconcile(ctx)
}

// GetGVisorInstallationManagedResourceName returns the name of the managed resource.
func GetGVisorInstallationManagedResourceName(cr *extensionsv1alpha1.ContainerRuntime) string {
	return fmt.Sprintf("%s-%s", GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name)
}

func withLocalObjectRefs(refs ...string) []corev1.LocalObjectReference {
	var localObjectRefs []corev1.LocalObjectReference
	for _, ref := range refs {
		localObjectRefs = append(localObjectRefs, corev1.LocalObjectReference{Name: ref})
	}
	return localObjectRefs
}

func buildSecret(cl client.Client, gVisorConfig []byte, namespace, secretName string) (*builder.Secret, []corev1.LocalObjectReference) {
	return builder.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.GVisorConfigKey: gVisorConfig}).
		WithNamespacedName(namespace, secretName), withLocalObjectRefs(secretName)
}
