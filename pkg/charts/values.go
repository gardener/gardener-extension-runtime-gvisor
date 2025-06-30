// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	"fmt"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-extension-runtime-gvisor/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/imagevector"
	gvisorconfiguration "github.com/gardener/gardener-extension-runtime-gvisor/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

// GVisorConfigKey is the key for the gVisor configuration.
const GVisorConfigKey = "config.yaml"

var decoder runtime.Decoder

func init() {
	scheme := runtime.NewScheme()
	runtimeutils.Must(gvisorconfiguration.AddToScheme(scheme))
	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
}

// RenderGVisorInstallationChart renders the gVisor installation chart
func RenderGVisorInstallationChart(renderer chartrenderer.Interface, cr *extensionsv1alpha1.ContainerRuntime) ([]byte, error) {

	providerConfig := &gvisorconfiguration.GVisorConfiguration{}
	if cr.Spec.ProviderConfig != nil {
		if _, _, err := decoder.Decode(cr.Spec.ProviderConfig.Raw, nil, providerConfig); err != nil {
			// TODO: Add admission component and move validation there by using strict decoding, for example: https://github.com/gardener/gardener-extension-provider-aws/pull/307.
			return nil, v1beta1helper.NewErrorWithCodes(fmt.Errorf("could not decode provider config: %w", err), gardencorev1beta1.ErrorConfigurationProblem)
		}
	}

	runscConfigFlags := ""
	if providerConfig.ConfigFlags != nil && len(*providerConfig.ConfigFlags) > 0 {
		for key, value := range *providerConfig.ConfigFlags {
			// the API allows to set arbitrary flags, but we only allow the following flags for now
			// A list of all supported flags can be found here: https://github.com/google/gvisor/blob/master/runsc/config/flags.go
			if key == "net-raw" && (value == "true" || value == "false") {
				runscConfigFlags += fmt.Sprintf("%s = \"%s\"\n", key, value)
			}
			if key == "debug" && value == "true" {
				runscConfigFlags += fmt.Sprintf("%s = \"%s\"\n", key, "true")
				runscConfigFlags += "debug-log = \"/var/log/runsc/%ID%/gvisor-%COMMAND%.log\"\n"
			}
			if key == "nvproxy" && value == "true" {
				runscConfigFlags += fmt.Sprintf("%s = \"%s\"\n", key, "true")
			}
		}
	}

	nodeSelectorValue := map[string]string{
		extensionsv1alpha1.CRINameWorkerLabel: string(extensionsv1alpha1.CRINameContainerD),
	}

	for key, value := range cr.Spec.WorkerPool.Selector.MatchLabels {
		nodeSelectorValue[key] = value
	}

	configChartValues := map[string]interface{}{
		"binFolder":    cr.Spec.BinaryPath,
		"nodeSelector": nodeSelectorValue,
		"workergroup":  cr.Spec.WorkerPool.Name,
		"configFlags":  runscConfigFlags,
	}

	gvisorChartValues := map[string]interface{}{
		"config": configChartValues,
		"images": map[string]string{
			gvisor.RuntimeGVisorInstallationImageName: imagevector.FindImage(gvisor.RuntimeGVisorInstallationImageName),
		},
	}

	release, err := renderer.RenderEmbeddedFS(charts.InternalChart, gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gvisorChartValues)
	if err != nil {
		return nil, err
	}
	return release.Manifest(), nil
}

// RenderGVisorChart renders the gVisor chart
func RenderGVisorChart(renderer chartrenderer.Interface) ([]byte, error) {
	gvisorChartValues := map[string]interface{}{}

	release, err := renderer.RenderEmbeddedFS(charts.InternalChart, gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gvisorChartValues)
	if err != nil {
		return nil, err
	}
	return release.Manifest(), nil
}
