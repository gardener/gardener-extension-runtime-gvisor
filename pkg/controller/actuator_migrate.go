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

// Migrate implements ContainerRuntime.Actuator.
func (a *actuator) Migrate(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.ContainerRuntime, _ *extensionscontroller.Cluster) error {
	return nil
}
