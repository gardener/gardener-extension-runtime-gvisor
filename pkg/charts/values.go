// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	"fmt"

	"github.com/gardener/gardener-extension-runtime-gvisor/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/imagevector"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"

	gvisorconfiguration "github.com/gardener/gardener-extension-runtime-gvisor/pkg/apis/config/v1alpha1"
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

	providerConfig := &gvisorconfiguration.GvisorConfiguration{}
	if cr.Spec.ProviderConfig != nil {
		if _, _, err := decoder.Decode(cr.Spec.ProviderConfig.Raw, nil, providerConfig); err != nil {
			return nil, fmt.Errorf("could not decode provider config: %w", err)
		}
	}

	runscConfig := ""
	if providerConfig.AdditionalCapabilities != nil {
		if providerConfig.AdditionalCapabilities.CapabilityNetRaw != nil && *providerConfig.AdditionalCapabilities.CapabilityNetRaw {
			runscConfig += "net-raw = \"true\"\n"
		}
		if providerConfig.AdditionalCapabilities.CapabilitySysAdmin != nil && *providerConfig.AdditionalCapabilities.CapabilitySysAdmin {
			runscConfig += "sys-admin = \"true\"\n"
		}
	}

	nodeSelectorValue := map[string]string{
		extensionsv1alpha1.CRINameWorkerLabel: string(extensionsv1alpha1.CRINameContainerD),
	}

	for key, value := range cr.Spec.WorkerPool.Selector.MatchLabels {
		nodeSelectorValue[key] = value
	}

	configChartValues := map[string]interface{}{
		"binFolder":              cr.Spec.BinaryPath,
		"nodeSelector":           nodeSelectorValue,
		"workergroup":            cr.Spec.WorkerPool.Name,
		"additionalCapabilities": runscConfig,
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
