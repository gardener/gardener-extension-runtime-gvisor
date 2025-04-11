// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

// Delete implements ContainerRuntime.Actuator.
func (a *actuator) Delete(ctx context.Context, log logr.Logger, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	var (
		installationManagedResourceName = GVisorInstallationManagedResourceName + "-" + cr.Spec.WorkerPool.Name
		forceDelete                     = cluster != nil && v1beta1helper.ShootNeedsForceDeletion(cluster.Shoot)
	)

	log.Info("Deleting managed resource due to the deletion of the corresponding ContainerRuntime", "managedResourceName", installationManagedResourceName)
	if err := a.deleteManagedResource(ctx, cr.Namespace, installationManagedResourceName, forceDelete); err != nil {
		return err
	}

	// delete the gVisor managed resource if all ContainerRuntime CRDs of type gVisor have a deletion timestamp
	list := &extensionsv1alpha1.ContainerRuntimeList{}
	if err := a.client.List(ctx, list, client.InNamespace(cr.Namespace)); err != nil {
		return err
	}

	if isGVisorInstallationRequired(cr.Name, list) {
		log.Info("gVisor is still required in the cluster - go ahead with ContainerRuntime deletion")
		return nil
	}
	log.Info("Deleting managed resource - no worker pool in the Shoot cluster requires gVisor any more", "managedResourceName", GVisorManagedResourceName)

	return a.deleteManagedResource(ctx, cr.Namespace, GVisorManagedResourceName, forceDelete)
}

func isGVisorInstallationRequired(name string, list *extensionsv1alpha1.ContainerRuntimeList) bool {
	for _, cr := range list.Items {
		if cr.Name != name && cr.Spec.Type == gvisor.Type && cr.DeletionTimestamp == nil {
			return true
		}
	}
	return false
}

// ForceDelete implements ContainerRuntime.Actuator.
func (a *actuator) ForceDelete(ctx context.Context, log logr.Logger, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	return a.Delete(ctx, log, cr, cluster)
}
