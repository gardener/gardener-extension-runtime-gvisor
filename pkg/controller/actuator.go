// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/containerruntime"
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
