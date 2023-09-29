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

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
)

const (
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
	gVisorChart, err := charts.RenderGVisorChart(chartRenderer, pspDisabled)
	if err != nil {
		return err
	}

	if err := managedresources.CreateForShoot(ctx, a.client, cr.Namespace, GVisorManagedResourceName, "extension-runtime-gvisor", false, map[string][]byte{charts.GVisorConfigKey: gVisorChart}); err != nil {
		return err
	}

	log.Info("Installing gVisor", "shoot", cluster.Shoot.Name, "shootNamespace", cluster.Shoot.Namespace, "workerPoolName", cr.Spec.WorkerPool.Name)
	gVisorInstallationChart, err := charts.RenderGVisorInstallationChart(chartRenderer, cr)
	if err != nil {
		return err
	}

	installSecretName := fmt.Sprintf("%s-%s", GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name)
	secretName, secret := managedresources.NewSecret(a.client, cr.Namespace, installSecretName, map[string][]byte{charts.GVisorConfigKey: gVisorInstallationChart}, true)
	installMRName := fmt.Sprintf("%s-%s", GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name)
	managedResource := managedresources.NewForShoot(a.client, cr.Namespace, installMRName, "extension-runtime-gvisor", false).WithSecretRef(secretName)

	if err := secret.Reconcile(ctx); err != nil {
		return err
	}
	return managedResource.Reconcile(ctx)
}
