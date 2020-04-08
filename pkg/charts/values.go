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

package charts

import (
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/imagevector"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GVisorConfigKey = "config.yaml"

// RenderGVisorInstallationChart renders the gVisor installation chart
func RenderGVisorInstallationChart(renderer chartrenderer.Interface, cr *extensionsv1alpha1.ContainerRuntime) ([]byte, error) {
	nodeSelectorValue := map[string]string{
		extensionsv1alpha1.CRINameWorkerLabel: extensionsv1alpha1.CRINameContainerD,
	}

	for key, value := range cr.Spec.WorkerPool.Selector.MatchLabels {
		nodeSelectorValue[key] = value
	}

	configChartValues := map[string]interface{}{
		"binFolder":    cr.Spec.BinaryPath,
		"nodeSelector": nodeSelectorValue,
		"workergroup":  cr.Spec.WorkerPool.Name,
	}

	gvisorChartValues := map[string]interface{}{
		"config": configChartValues,
		"images": map[string]string{
			gvisor.RuntimeGVisorInstallationImageName: imagevector.FindImage(gvisor.RuntimeGVisorInstallationImageName),
		},
	}

	release, err := renderer.Render(gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gvisorChartValues)
	if err != nil {
		return nil, err
	}
	return release.Manifest(), nil
}

// RenderGVisorChart renders the gVisor chart
func RenderGVisorChart(renderer chartrenderer.Interface, kubernetesVersion string) ([]byte, error) {
	configChartValues := map[string]interface{}{
		"kubernetesVersion": kubernetesVersion,
	}

	gvisorChartValues := map[string]interface{}{
		"config": configChartValues,
	}

	release, err := renderer.Render(gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gvisorChartValues)
	if err != nil {
		return nil, err
	}
	return release.Manifest(), nil
}
