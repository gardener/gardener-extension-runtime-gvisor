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

package charts_test

import (
	"fmt"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	mockchartrenderer "github.com/gardener/gardener/pkg/chartrenderer/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/manifest"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/imagevector"
)

var _ = Describe("Chart package test", func() {
	Describe("#RenderGvisorChart", func() {
		var (
			ctrl              *gomock.Controller
			mockChartRenderer *mockchartrenderer.MockInterface

			testManifestContent = "test-content"
			mkManifest          = func(name string) manifest.Manifest {
				return manifest.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
			workerGroup = "worker-gvisor"

			cr = extensionsv1alpha1.ContainerRuntime{
				Spec: extensionsv1alpha1.ContainerRuntimeSpec{
					BinaryPath: "/path/test",
					WorkerPool: extensionsv1alpha1.ContainerRuntimeWorkerPool{
						Name: workerGroup,
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{"worker.gardener.cloud/pool": "gvisor-pool"},
						},
					},
					DefaultSpec: extensionsv1alpha1.DefaultSpec{
						Type: "type"},
				},
			}
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockChartRenderer = mockchartrenderer.NewMockInterface(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("Render Gvisor chart correctly", func() {
			renderedValues := map[string]interface{}{
				"pspDisabled": true,
			}

			mockChartRenderer.EXPECT().Render(gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gomock.Eq(renderedValues)).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []manifest.Manifest{
					mkManifest(charts.GVisorConfigKey),
				},
			}, nil)

			_, err := charts.RenderGVisorChart(mockChartRenderer, true)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Render Gvisor installation chart correctly", func() {
			renderedValues := map[string]interface{}{
				"images": map[string]string{
					"runtime-gvisor-installation": imagevector.FindImage(gvisor.RuntimeGVisorInstallationImageName),
				},
				"config": map[string]interface{}{
					"nodeSelector": map[string]string{
						extensionsv1alpha1.CRINameWorkerLabel: string(extensionsv1alpha1.CRINameContainerD),
						"worker.gardener.cloud/pool":          "gvisor-pool",
					},
					"binFolder":   "/path/test",
					"workergroup": workerGroup,
				},
			}

			mockChartRenderer.EXPECT().Render(gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gomock.Eq(renderedValues)).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []manifest.Manifest{
					mkManifest(charts.GVisorConfigKey),
				},
			}, nil)

			_, err := charts.RenderGVisorInstallationChart(mockChartRenderer, &cr)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
