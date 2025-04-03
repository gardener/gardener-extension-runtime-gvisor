// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
)

// Migrate implements ContainerRuntime.Actuator.
func (a *actuator) Migrate(ctx context.Context, log logr.Logger, cr *extensionsv1alpha1.ContainerRuntime, _ *extensionscontroller.Cluster) error {
	managedResourceName := GVisorInstallationManagedResourceName + "-" + cr.Spec.WorkerPool.Name

	log.Info("Setting keepObjects=true for managed resource due to the migration of the corresponding ContainerRuntime", "managedResourceName", managedResourceName, "namespace", cr.Namespace, "containerRuntime", cr.Name)
	if err := managedresources.SetKeepObjects(ctx, a.client, cr.Namespace, managedResourceName, true); err != nil {
		return fmt.Errorf("could not keep objects of managed resource %q: %w", managedResourceName, err)
	}

	log.Info("Deleting managed resource due to the migration of the corresponding ContainerRuntime", "managedResourceName", managedResourceName, "namespace", cr.Namespace, "containerRuntime", cr.Name)
	if err := a.deleteManagedResource(ctx, cr.Namespace, managedResourceName, false); err != nil {
		return fmt.Errorf("could not delete managed resource %q: %w", managedResourceName, err)
	}

	// We can directly set `keepObjects=true` and delete the GVisor ManagedResource because all ContainerRuntimes are migrated
	// during control plane migration. If the GVisor ManagedResource was already deleted no error is returned.
	log.Info("Setting keepObjects=true as part of the migration operation", "managedResourceName", GVisorManagedResourceName)
	if err := managedresources.SetKeepObjects(ctx, a.client, cr.Namespace, GVisorManagedResourceName, true); err != nil {
		return fmt.Errorf("could not keep objects of managed resource %q: %w", GVisorManagedResourceName, err)
	}
	log.Info("Deleting managed resource as part of the migration operation", "managedResourceName", GVisorManagedResourceName)
	if err := a.deleteManagedResource(ctx, cr.Namespace, GVisorManagedResourceName, false); err != nil {
		return fmt.Errorf("could not delete managed resource %q: %w", GVisorManagedResourceName, err)
	}

	return nil
}
