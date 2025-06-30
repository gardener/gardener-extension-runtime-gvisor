// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts_test

import (
	"errors"
	"fmt"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	mockchartrenderer "github.com/gardener/gardener/pkg/chartrenderer/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"helm.sh/helm/v3/pkg/releaseutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"

	internalcharts "github.com/gardener/gardener-extension-runtime-gvisor/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/imagevector"
	gvisorconfiguration "github.com/gardener/gardener-extension-runtime-gvisor/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

var _ = Describe("Chart package test", func() {
	Describe("#RenderGvisorChart", func() {
		var (
			ctrl               *gomock.Controller
			mockChartRenderer  *mockchartrenderer.MockInterface
			expectedHelmValues map[string]interface{}

			testManifestContent = "test-content"
			mkManifest          = func(name string) releaseutil.Manifest {
				return releaseutil.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
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
						Type: "type",
					},
				},
			}
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockChartRenderer = mockchartrenderer.NewMockInterface(ctrl)
			expectedHelmValues = map[string]interface{}{
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
					"configFlags": "",
				},
			}

		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("Render Gvisor chart correctly", func() {
			renderedValues := map[string]interface{}{}

			mockChartRenderer.EXPECT().RenderEmbeddedFS(internalcharts.InternalChart, gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gomock.Eq(renderedValues)).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []releaseutil.Manifest{
					mkManifest(charts.GVisorConfigKey),
				},
			}, nil)

			_, err := charts.RenderGVisorChart(mockChartRenderer)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Render Gvisor installation chart correctly with default settings", func() {
			mockChartRenderer.EXPECT().RenderEmbeddedFS(internalcharts.InternalChart, gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gomock.Eq(expectedHelmValues)).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []releaseutil.Manifest{
					mkManifest(charts.GVisorConfigKey),
				},
			}, nil)

			_, err := charts.RenderGVisorInstallationChart(mockChartRenderer, &cr)
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("Provider config decoding",
			func(providerConfig *gvisorconfiguration.GVisorConfiguration, expectedError string) {
				rawJson, _ := json.Marshal(providerConfig)

				cr.Spec.ProviderConfig = &runtime.RawExtension{Raw: rawJson}

				_, err := charts.RenderGVisorInstallationChart(mockChartRenderer, &cr)
				var coder helper.Coder
				Expect(errors.As(err, &coder)).To(Equal(true))
				codes := coder.Codes()
				Expect(len(codes)).To(Equal(1))
				Expect(codes[0]).To(Equal(gardencorev1beta1.ErrorCode("ERR_CONFIGURATION_PROBLEM")))
				Expect(err.Error()).To(ContainSubstring(expectedError))
			},
			Entry("fail on invalid API",
				&gvisorconfiguration.GVisorConfiguration{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "some.group.com/v1alpha1",
						Kind:       "GVisorConfiguration",
					},
					ConfigFlags: &map[string]string{"any": "flag"},
				},
				`no kind "GVisorConfiguration" is registered for version "some.group.com/v1alpha1" in scheme`,
			),
			Entry("fail on unsupported API version",
				&gvisorconfiguration.GVisorConfiguration{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gvisorconfiguration.GroupName + "/v1alpha2",
						Kind:       "GVisorConfiguration",
					},
					ConfigFlags: &map[string]string{"any": "flag"},
				},
				`no kind "GVisorConfiguration" is registered for version "gvisor.runtime.extensions.config.gardener.cloud/v1alpha2" in scheme`,
			),
			Entry("fail on invalid Kind",
				&gvisorconfiguration.GVisorConfiguration{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gvisorconfiguration.GroupName + "/v1alpha1",
						Kind:       "ControllerConfiguration",
					},
					ConfigFlags: &map[string]string{"any": "flag"},
				},
				"converting (v1alpha1.ControllerConfiguration) to (v1alpha1.GVisorConfiguration): unknown conversion",
			),
		)

		DescribeTable("Render Gvisor installation chart correctly with valid provider config",
			func(configFlags map[string]string, expectedConfigFlags string) {
				providerConfig := &gvisorconfiguration.GVisorConfiguration{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gvisorconfiguration.GroupName + "/v1alpha1",
						Kind:       "GVisorConfiguration",
					},
					ConfigFlags: &configFlags,
				}

				// this function acts as a little self test on schema encoding that is currently not used in production code.
				rawJson, err := encodeProviderConfig(providerConfig)

				// this indicates that either the provider config is incorrect
				// or the scheme is not registered correctly
				Expect(err).NotTo(HaveOccurred())

				cr.Spec.ProviderConfig = &runtime.RawExtension{Raw: rawJson}

				// provider config capabilities should be rendered into values
				expectedHelmValues["config"].(map[string]interface{})["configFlags"] = expectedConfigFlags

				mockChartRenderer.EXPECT().RenderEmbeddedFS(internalcharts.InternalChart, gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gomock.Eq(expectedHelmValues)).Return(&chartrenderer.RenderedChart{
					ChartName: "test",
					Manifests: []releaseutil.Manifest{
						mkManifest(charts.GVisorConfigKey),
					},
				}, nil)

				_, err = charts.RenderGVisorInstallationChart(mockChartRenderer, &cr)
				Expect(err).NotTo(HaveOccurred())
			},
			Entry("no-flags", map[string]string{}, ""),
			Entry("nvproxy-flag", map[string]string{"nvproxy": "true"}, "nvproxy = \"true\"\n"),
			Entry("net-raw-flag", map[string]string{"net-raw": "true"}, "net-raw = \"true\"\n"),
			Entry("debug-flag",
				map[string]string{"debug": "true"},
				"debug = \"true\"\ndebug-log = \"/var/log/runsc/%ID%/gvisor-%COMMAND%.log\"\n"),
		)
	})
})

// helper function to encode the provider config to a raw JSON byte array
// Note that this function will not be used in production code, but is a self test
// It uses only the Go Type system to encode the provider config
// the API version and Kind are not checked
func encodeProviderConfig(providerConfig *gvisorconfiguration.GVisorConfiguration) ([]byte, error) {
	scheme := runtime.NewScheme()
	runtimeutils.Must(gvisorconfiguration.AddToScheme(scheme))
	codecFactory := serializer.NewCodecFactory(scheme)
	encoder := codecFactory.LegacyCodec(gvisorconfiguration.SchemeGroupVersion)

	encoded, err := runtime.Encode(encoder, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("could not encode provider config: %w", err)
	}

	return encoded, nil
}
