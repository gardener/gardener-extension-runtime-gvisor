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

package controller_test

import (
	"context"
	resourcemanagerv1alpha1 "github.com/gardener/gardener-resource-manager/pkg/apis/resources/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/controller"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockextensionscontroller "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller"
	mockchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener/chartrenderer"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/manifest"
)

const (
	shootVersion = "1.14.0"
)

var _ = Describe("Chart package test", func() {

	Describe("#Actuator", func() {
		var (
			ctrl            = gomock.NewController(GinkgoT())
			crf             = mockextensionscontroller.NewMockChartRendererFactory(ctrl)
			client          = mockclient.NewMockClient(ctrl)
			ctx             = context.TODO()
			chartName       = "chartName"
			manifestContent = "manifestContent"
			namespaceName   = "namespace"
			cr              = &extensionsv1alpha1.ContainerRuntime{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName},
				Spec: extensionsv1alpha1.ContainerRuntimeSpec{
					BinaryPath: "/path/test",
					DefaultSpec: extensionsv1alpha1.DefaultSpec{
						Type: "type"}}}
			cluster = &extensionscontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					Spec: gardencorev1beta1.ShootSpec{
						Kubernetes: gardencorev1beta1.Kubernetes{
							Version: shootVersion,
						},
					},
				},
			}
			errNotFound = &errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
		)

		// Create actuator
		a := controller.NewActuator(crf)
		client = mockclient.NewMockClient(ctrl)
		a.(inject.Client).InjectClient(client)

		It("Reconcile correctly", func() {

			// Create mock chart renderer and factory
			chartRenderer := mockchartrenderer.NewMockInterface(ctrl)
			crf.EXPECT().NewChartRendererForShoot(shootVersion).Return(chartRenderer, nil)
			renderedChart := &chartrenderer.RenderedChart{
				ChartName: chartName,
				Manifests: []manifest.Manifest{{Content: manifestContent}},
			}
			chartRenderer.EXPECT().Render(gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(renderedChart, nil)

			errNotFound := &errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}

			// Validate deployed secret
			deployedSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorRuntimeSecretName, Namespace: namespaceName},
				Data:       map[string][]byte{charts.GVisorConfigKey: renderedChart.Manifest()},
				Type:       corev1.SecretTypeOpaque,
			}
			client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			client.EXPECT().Create(ctx, deployedSecret).Return(nil)

			// Validate deployed managed resource
			managedResource := &resourcemanagerv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorRuntimeSecretName, Namespace: namespaceName},
				Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
					SecretRefs: []corev1.LocalObjectReference{
						{Name: controller.GVisorRuntimeSecretName},
					},
					InjectLabels: map[string]string{extensionscontroller.ShootNoCleanupLabel: "true"},
				},
			}
			client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			client.EXPECT().Create(ctx, managedResource).Return(nil)

			// Create mock client
			err := a.Reconcile(ctx, cr, cluster)

			//client.EXPECT().Create()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete correctly", func() {

			// Validate deleted secret
			deployedSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorRuntimeSecretName, Namespace: namespaceName},
			}
			//client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			client.EXPECT().Delete(ctx, deployedSecret).Return(nil)

			// Validate deployed managed resource
			managedResource := &resourcemanagerv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorRuntimeSecretName, Namespace: namespaceName},
			}
			//client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			client.EXPECT().Delete(ctx, managedResource).Return(nil)

			// Create mock client
			err := a.Delete(ctx, cr, cluster)

			//client.EXPECT().Delete()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
