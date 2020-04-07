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
	"time"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"

	resourcemanager "github.com/gardener/gardener-resource-manager/pkg/manager"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Delete implements ContainerRuntime.Actuator.
func (a *actuator) Delete(ctx context.Context, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	a.logger.Info("Deleting managed resource due to the deletion of the corresponding ContainerRuntime", "managed resource", fmt.Sprintf("%s-%s", GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name), "namespace", cr.Namespace, "containerRuntime", cr.Name)
	if err := a.deleteManagedResource(ctx, cr.Namespace, fmt.Sprintf("%s-%s", GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name), fmt.Sprintf("%s-%s", GVisorInstallationSecretName, cr.Spec.WorkerPool.Name)); err != nil {
		return err
	}

	// delete the gVisor managed resource if all ContainerRuntime CRDs of type gVisor have a deletion timestamp
	list := &extensionsv1alpha1.ContainerRuntimeList{}
	if err := a.client.List(ctx, list, client.InNamespace(cr.Namespace)); err != nil {
		return err
	}

	if isGVisorInstallationRequired(cr, list) {
		a.logger.Info("gVisor is still required in the cluster - go ahead with ContainerRuntime deletion", "namespace", cr.Namespace, "containerRuntime", cr.Name)
		return nil
	}
	a.logger.Info("Deleting managed resource - no worker pool in the Shoot cluster requires gVisor any more", "managed resource", GVisorManagedResourceName)

	return a.deleteManagedResource(ctx, cr.Namespace, GVisorManagedResourceName, GVisorSecretName)
}

func (a *actuator) deleteManagedResource(ctx context.Context, namespace, managedResourceName, secretName string) error {
	if err := resourcemanager.
		NewManagedResource(a.client).
		WithNamespacedName(namespace, managedResourceName).
		Delete(ctx); err != nil {
		return err
	}

	if err := resourcemanager.
		NewSecret(a.client).
		WithNamespacedName(namespace, secretName).
		Delete(ctx); err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := extensionscontroller.WaitUntilManagedResourceDeleted(timeoutCtx, a.client, namespace, managedResourceName); err != nil {
		return err
	}
	return nil
}

func isGVisorInstallationRequired(c *extensionsv1alpha1.ContainerRuntime, list *extensionsv1alpha1.ContainerRuntimeList) bool {
	for _, cr := range list.Items {
		if cr.Name != c.Name && cr.Spec.DefaultSpec.Type == gvisor.Type && cr.DeletionTimestamp == nil {
			return true
		}
	}
	return false
}
