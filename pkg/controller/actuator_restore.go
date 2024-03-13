// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
)

// Restore implements ContainerRuntime.Actuator.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	return a.Reconcile(ctx, log, cr, cluster)
}
