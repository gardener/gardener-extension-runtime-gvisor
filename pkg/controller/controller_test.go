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
	"fmt"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/controller"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/containerruntime"
	mockextensionscontroller "github.com/gardener/gardener/extensions/pkg/controller/mock"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	mockchartrenderer "github.com/gardener/gardener/pkg/chartrenderer/mock"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/helm/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	shootVersion = "1.24.0"
)

var _ = Describe("Chart package test", func() {

	Describe("#Actuator", func() {
		var (
			ctrl              *gomock.Controller
			crf               *mockextensionscontroller.MockChartRendererFactory
			mockChartRenderer *mockchartrenderer.MockInterface
			mockClient        *mockclient.MockClient
			a                 containerruntime.Actuator

			ctx = context.TODO()
			log = logf.Log.WithName("test")

			chartName       = "chartName"
			manifestContent = "manifestContent"
			namespaceName   = "namespace"
			workerGroup     = "worker-gvisor"
			cr              = &extensionsv1alpha1.ContainerRuntime{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: "test-cr"},
				Spec: extensionsv1alpha1.ContainerRuntimeSpec{
					BinaryPath: "/path/test",
					WorkerPool: extensionsv1alpha1.ContainerRuntimeWorkerPool{
						Name: workerGroup,
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{"worker.gardener.cloud/pool": "gvisor-pool"},
						},
					},
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

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			crf = mockextensionscontroller.NewMockChartRendererFactory(ctrl)
			mockChartRenderer = mockchartrenderer.NewMockInterface(ctrl)
			mockClient = mockclient.NewMockClient(ctrl)
			a = controller.NewActuator(crf)

			err := a.(inject.Client).InjectClient(mockClient)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("Reconcile correctly - first ContainerRuntime", func() {
			// Create mock chart renderer and factory
			chartRenderer := mockchartrenderer.NewMockInterface(ctrl)
			crf.EXPECT().NewChartRendererForShoot(shootVersion).Return(chartRenderer, nil)
			renderedChart := &chartrenderer.RenderedChart{
				ChartName: chartName,
				Manifests: []manifest.Manifest{{Content: manifestContent}},
			}

			// ---------- gVisor Preparation -------------------
			// chart renderer renders chart with path "gvisor"
			chartRenderer.EXPECT().Render(gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(renderedChart, nil)
			// mockClient creates or update secret for managed resource "extension-runtime-gvisor"
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorSecretName, Namespace: namespaceName},
				Data:       map[string][]byte{charts.GVisorConfigKey: renderedChart.Manifest()},
				Type:       corev1.SecretTypeOpaque,
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, secret).Return(nil)
			// Validate deployed managed resource "extension-runtime-gvisor"
			managedResource := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorManagedResourceName, Namespace: namespaceName},
				Spec: resourcesv1alpha1.ManagedResourceSpec{
					SecretRefs: []corev1.LocalObjectReference{
						{Name: controller.GVisorSecretName},
					},
				},
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, managedResource).Return(nil)

			// ---------- gVisor Installation -------------------
			chartRenderer.EXPECT().Render(gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(renderedChart, nil)
			// Validate deployed secret
			installationSecretName := fmt.Sprintf("%s-%s", controller.GVisorInstallationSecretName, cr.Spec.WorkerPool.Name)
			installationSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: installationSecretName, Namespace: namespaceName},
				Data:       map[string][]byte{charts.GVisorConfigKey: renderedChart.Manifest()},
				Type:       corev1.SecretTypeOpaque,
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, installationSecret).Return(nil)

			// Validate deployed managed resource
			installationMResourceName := fmt.Sprintf("%s-%s", controller.GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name)
			managedResourceInstallation := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: installationMResourceName, Namespace: namespaceName},
				Spec: resourcesv1alpha1.ManagedResourceSpec{
					SecretRefs: []corev1.LocalObjectReference{
						{Name: installationSecretName},
					},
				},
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, managedResourceInstallation).Return(nil)

			// Create mock mockClient
			err := a.Reconcile(ctx, log, cr, cluster)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Reconcile correctly - add additional worker pool with gVisor", func() {
			crf.EXPECT().NewChartRendererForShoot(shootVersion).Return(mockChartRenderer, nil)
			renderedChart := &chartrenderer.RenderedChart{
				ChartName: chartName,
				Manifests: []manifest.Manifest{{Content: manifestContent}},
			}

			// ---------- gVisor Preparation -------------------
			// chart renderer renders chart with path "gvisor"
			mockChartRenderer.EXPECT().Render(gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(renderedChart, nil)
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorSecretName, Namespace: namespaceName},
				Data:       map[string][]byte{charts.GVisorConfigKey: renderedChart.Manifest()},
				Type:       corev1.SecretTypeOpaque,
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, secret).Return(nil)
			// Validate deployed managed resource "extension-runtime-gvisor"
			managedResource := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorManagedResourceName, Namespace: namespaceName},
				Spec: resourcesv1alpha1.ManagedResourceSpec{
					SecretRefs: []corev1.LocalObjectReference{
						{Name: controller.GVisorSecretName},
					},
				},
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, managedResource).Return(nil)

			// ---------- gVisor Installation -------------------
			mockChartRenderer.EXPECT().Render(gvisor.InstallationChartPath, gvisor.InstallationReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(renderedChart, nil)
			// Validate deployed secret
			installationSecretName := fmt.Sprintf("%s-%s", controller.GVisorInstallationSecretName, cr.Spec.WorkerPool.Name)
			installationSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: installationSecretName, Namespace: namespaceName},
				Data:       map[string][]byte{charts.GVisorConfigKey: renderedChart.Manifest()},
				Type:       corev1.SecretTypeOpaque,
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, installationSecret).Return(nil)

			// Validate deployed managed resource
			installationMResourceName := fmt.Sprintf("%s-%s", controller.GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name)
			managedResourceInstallation := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: installationMResourceName, Namespace: namespaceName},
				Spec: resourcesv1alpha1.ManagedResourceSpec{
					SecretRefs: []corev1.LocalObjectReference{
						{Name: installationSecretName},
					},
				},
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)
			mockClient.EXPECT().Create(ctx, managedResourceInstallation).Return(nil)

			// Create mock mockClient
			err := a.Reconcile(ctx, log, cr, cluster)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete correctly - shoot does not require gVisor any more", func() {

			// ---------- Deletion of GVisor Installation -------------------

			// Validate deployed managed resource
			installationManagedResource := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", controller.GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name), Namespace: namespaceName},
			}
			mockClient.EXPECT().Delete(ctx, installationManagedResource).Return(nil)

			// Validate deleted secret
			installationSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", controller.GVisorInstallationSecretName, cr.Spec.WorkerPool.Name), Namespace: namespaceName},
			}
			mockClient.EXPECT().Delete(ctx, installationSecret).Return(nil)
			// wait for managed resource to be deleted
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)

			// ---------- Deletion of GVisor Prerequisites -------------------

			mockClient.EXPECT().List(context.TODO(), gomock.AssignableToTypeOf(&extensionsv1alpha1.ContainerRuntimeList{}), gomock.Any()).DoAndReturn(func(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
				Expect(list).To(BeAssignableToTypeOf(&extensionsv1alpha1.ContainerRuntimeList{}))
				now := metav1.Now()
				list.(*extensionsv1alpha1.ContainerRuntimeList).Items = []extensionsv1alpha1.ContainerRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now},
						Spec: extensionsv1alpha1.ContainerRuntimeSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{Type: gvisor.Type},
						},
					},
					{
						Spec: extensionsv1alpha1.ContainerRuntimeSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{Type: "kata"},
						},
					},
					*cr,
				}
				return nil
			})

			// Validate deployed managed resource
			managedResource := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorManagedResourceName, Namespace: namespaceName},
			}
			mockClient.EXPECT().Delete(ctx, managedResource).Return(nil)

			// delete secret of managed resource
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: controller.GVisorSecretName, Namespace: namespaceName},
			}
			mockClient.EXPECT().Delete(ctx, secret).Return(nil)

			// wait for managed resource to be deleted
			managedResourceStillAvailable := func(ctx context.Context, key client.ObjectKey, obj runtime.Object, _ ...client.GetOption) error {
				Expect(obj).To(BeAssignableToTypeOf(&resourcesv1alpha1.ManagedResource{}))
				object := client.ObjectKeyFromObject(managedResource)
				Expect(key).To(Equal(object))
				now := metav1.Now()
				obj.(*resourcesv1alpha1.ManagedResource).ObjectMeta.DeletionTimestamp = &now
				return nil
			}
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(managedResourceStillAvailable).Times(4)
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)

			// Create mock mockClient
			err := a.Delete(ctx, log, cr, cluster)

			//mockClient.EXPECT().Delete()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete correctly - another worker pool still requires gVisor.", func() {

			// ---------- Deletion of GVisor Installation -------------------

			// Validate deployed managed resource
			installationManagedResource := &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", controller.GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name), Namespace: namespaceName},
			}
			mockClient.EXPECT().Delete(ctx, installationManagedResource).Return(nil)

			// Validate deleted secret
			installationSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", controller.GVisorInstallationSecretName, cr.Spec.WorkerPool.Name), Namespace: namespaceName},
			}
			mockClient.EXPECT().Delete(ctx, installationSecret).Return(nil)
			// wait for managed resource to be deleted
			mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(errNotFound)

			// ---------- Deletion of GVisor Prerequisites -------------------

			mockClient.EXPECT().List(context.TODO(), gomock.AssignableToTypeOf(&extensionsv1alpha1.ContainerRuntimeList{}), gomock.Any()).DoAndReturn(func(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
				Expect(list).To(BeAssignableToTypeOf(&extensionsv1alpha1.ContainerRuntimeList{}))
				list.(*extensionsv1alpha1.ContainerRuntimeList).Items = []extensionsv1alpha1.ContainerRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "abc",
						},
						Spec: extensionsv1alpha1.ContainerRuntimeSpec{
							WorkerPool:  extensionsv1alpha1.ContainerRuntimeWorkerPool{Name: "abcWorker"},
							DefaultSpec: extensionsv1alpha1.DefaultSpec{Type: gvisor.Type},
						},
					},
					*cr,
				}
				return nil
			})

			// Create mock mockClient
			err := a.Delete(ctx, log, cr, cluster)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
