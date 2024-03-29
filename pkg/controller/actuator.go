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
