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

	resourcemanagerv1alpha1 "github.com/gardener/gardener-resource-manager/api/resources/v1alpha1"
	"github.com/gardener/gardener-resource-manager/pkg/manager"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GVisorInstallationSecretName          = "extension-runtime-gvisor-installation"
	GVisorSecretName                      = "extension-runtime-gvisor"
	GVisorInstallationManagedResourceName = "extension-runtime-gvisor-installation"
	GVisorManagedResourceName             = "extension-runtime-gvisor"
)

// Reconcile implements ContainerRuntime.Actuator.
func (a *actuator) Reconcile(ctx context.Context, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("could not create chart renderer for shoot '%s', %w", cr.Namespace, err)
	}

	// if the managed resource containing the prerequisites is not available yet, create it.
	mr := &resourcemanagerv1alpha1.ManagedResource{}
	if err := a.client.Get(ctx, kutil.Key(cr.Namespace, GVisorManagedResourceName), mr); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		a.logger.Info("Preparing gVisor installation", "shoot", cluster.Shoot.Name, "namespace", cluster.Shoot.Namespace)
		// create MR containing the prerequisites for the installation DaemonSet
		gVisorChart, err := charts.RenderGVisorChart(chartRenderer, cluster.Shoot.Spec.Kubernetes.Version)
		if err != nil {
			return err
		}

		secret, secretRefs := gvisorSecret(a.client, gVisorChart, cr.Namespace, GVisorSecretName)
		if err := secret.Reconcile(ctx); err != nil {
			return err
		}

		if err := manager.
			NewManagedResource(a.client).
			WithNamespacedName(cr.Namespace, GVisorManagedResourceName).
			WithSecretRefs(secretRefs).
			Reconcile(ctx); err != nil {
			return fmt.Errorf("failed to create managed resource - prerequisite for the installation of gVisor, %w", err)
		}
	}

	a.logger.Info("Installing gVisor", "shoot", cluster.Shoot.Name, "namespace", cluster.Shoot.Namespace, "worker pool", cr.Spec.WorkerPool.Name)
	gVisorChart, err := charts.RenderGVisorInstallationChart(chartRenderer, cr)
	if err != nil {
		return err
	}

	secret, secretRefs := gvisorSecret(a.client, gVisorChart, cr.Namespace, fmt.Sprintf("%s-%s", GVisorInstallationSecretName, cr.Spec.WorkerPool.Name))
	if err := secret.Reconcile(ctx); err != nil {
		return err
	}

	return manager.
		NewManagedResource(a.client).
		WithNamespacedName(cr.Namespace, GetGVisorInstallationManagedResourceName(cr)).
		WithSecretRefs(secretRefs).
		Reconcile(ctx)
}

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

func gvisorSecret(cl client.Client, gVisorConfig []byte, namespace, secretName string) (*manager.Secret, []corev1.LocalObjectReference) {
	return manager.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.GVisorConfigKey: gVisorConfig}).
		WithNamespacedName(namespace, secretName), withLocalObjectRefs(secretName)
}
