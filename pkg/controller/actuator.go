// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/containerruntime"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type actuator struct {
	chartRendererFactory extensionscontroller.ChartRendererFactory

	client client.Client
}

// NewActuator creates a new Actuator that updates the status of the handled ContainerRuntime resources.
func NewActuator(c client.Client, chartRendererFactory extensionscontroller.ChartRendererFactory) containerruntime.Actuator {
	return &actuator{
		chartRendererFactory: chartRendererFactory,
		client:               c,
	}
}

func (a *actuator) deleteManagedResource(ctx context.Context, namespace, managedResourceName string, forceDelete bool) error {
	if err := managedresources.Delete(ctx, a.client, namespace, managedResourceName, true); err != nil {
		return err
	}

	if !forceDelete {
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		return managedresources.WaitUntilDeleted(timeoutCtx, a.client, namespace, managedResourceName)
	}

	return nil
}
