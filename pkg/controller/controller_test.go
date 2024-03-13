// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller_test

import (
	"context"
	"fmt"

	extensioncontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/containerruntime"
	"github.com/gardener/gardener/extensions/pkg/util"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/controller"
)

const (
	shootVersion = "1.28.0"
)

var _ = Describe("Controller tests", func() {
	Describe("#Actuator", func() {
		var (
			ctx context.Context
			c   client.Client

			managedResourceName   string
			managedResource       *resourcesv1alpha1.ManagedResource
			managedResourceSecret *corev1.Secret

			cr                           *extensionsv1alpha1.ContainerRuntime
			managedResourceInstallName   string
			managedResourceInstall       *resourcesv1alpha1.ManagedResource
			managedResourceInstallSecret *corev1.Secret

			cr2                           *extensionsv1alpha1.ContainerRuntime
			managedResourceInstall2Name   string
			managedResourceInstall2       *resourcesv1alpha1.ManagedResource
			managedResourceInstall2Secret *corev1.Secret

			a containerruntime.Actuator

			log = logf.Log.WithName("test")

			namespaceName = "namespace"
			workerGroup   = "worker-gvisor"

			cluster = &extensioncontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					Spec: gardencorev1beta1.ShootSpec{
						Kubernetes: gardencorev1beta1.Kubernetes{
							Version: shootVersion,
						},
					},
				},
			}
		)

		BeforeEach(func() {
			ctx = context.TODO()
			c = fake.NewClientBuilder().WithScheme(kubernetes.SeedScheme).Build()
			a = controller.NewActuator(c, extensioncontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot))

			managedResourceName = "extension-runtime-gvisor"
			managedResource = &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      managedResourceName,
					Namespace: namespaceName,
				},
			}
			managedResourceSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "managedresource-" + managedResource.Name,
					Namespace: namespaceName,
				},
			}

			cr = &extensionsv1alpha1.ContainerRuntime{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: "test-cr"},
				Spec: extensionsv1alpha1.ContainerRuntimeSpec{
					BinaryPath: "/path/test",
					WorkerPool: extensionsv1alpha1.ContainerRuntimeWorkerPool{
						Name: workerGroup,
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{"worker.gardener.cloud/pool": "gvisor-pool"},
						},
					},
					DefaultSpec: extensionsv1alpha1.DefaultSpec{Type: "gvisor"},
				},
			}
			managedResourceInstallName = fmt.Sprintf("extension-runtime-gvisor-installation-%s", cr.Spec.WorkerPool.Name)
			managedResourceInstall = &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      managedResourceInstallName,
					Namespace: namespaceName,
				},
			}
			managedResourceInstallSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "managedresource-" + managedResourceInstall.Name,
					Namespace: namespaceName,
				},
			}

			cr2 = &extensionsv1alpha1.ContainerRuntime{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: "test-cr-2"},
				Spec: extensionsv1alpha1.ContainerRuntimeSpec{
					BinaryPath: "/path/test",
					WorkerPool: extensionsv1alpha1.ContainerRuntimeWorkerPool{
						Name: workerGroup + "-2",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{"worker.gardener.cloud/pool": "gvisor-pool-2"},
						},
					},
					DefaultSpec: extensionsv1alpha1.DefaultSpec{Type: "gvisor"},
				},
			}
			managedResourceInstall2Name = fmt.Sprintf("extension-runtime-gvisor-installation-%s", cr2.Spec.WorkerPool.Name)
			managedResourceInstall2 = &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      managedResourceInstall2Name,
					Namespace: namespaceName,
				},
			}
			managedResourceInstall2Secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "managedresource-" + managedResourceInstall2.Name,
					Namespace: namespaceName,
				},
			}
		})

		deployOnSingleWorkerPool := func() {
			Expect(c.Create(ctx, cr)).To(Succeed())
			Expect(a.Reconcile(ctx, log, cr, cluster)).NotTo(HaveOccurred())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResource), managedResource)).To(Succeed())
			managedResourceSecret.Name = managedResource.Spec.SecretRefs[0].Name
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceSecret), managedResourceSecret)).To(Succeed())
			Expect(managedResourceSecret.Immutable).To(Equal(pointer.Bool(true)))
			Expect(managedResourceSecret.Data).To(HaveLen(1))

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall), managedResourceInstall)).To(Succeed())
			managedResourceInstallSecret.Name = managedResourceInstall.Spec.SecretRefs[0].Name
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstallSecret), managedResourceInstallSecret)).To(Succeed())
			Expect(managedResourceInstallSecret.Immutable).To(Equal(pointer.Bool(true)))
			Expect(managedResourceInstallSecret.Data).To(HaveLen(1))
		}

		deployOnTwoWorkerPools := func() {
			deployOnSingleWorkerPool()

			Expect(c.Create(ctx, cr2)).To(Succeed())
			Expect(a.Reconcile(ctx, log, cr2, cluster)).NotTo(HaveOccurred())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResource), managedResource)).To(Succeed())
			managedResourceSecret.Name = managedResource.Spec.SecretRefs[0].Name
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceSecret), managedResourceSecret)).To(Succeed())
			Expect(managedResourceSecret.Immutable).To(Equal(pointer.Bool(true)))
			Expect(managedResourceSecret.Data).To(HaveLen(1))

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall), managedResourceInstall)).To(Succeed())
			managedResourceInstallSecret.Name = managedResourceInstall.Spec.SecretRefs[0].Name
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstallSecret), managedResourceInstallSecret)).To(Succeed())
			Expect(managedResourceInstallSecret.Immutable).To(Equal(pointer.Bool(true)))
			Expect(managedResourceInstallSecret.Data).To(HaveLen(1))

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall2), managedResourceInstall2)).To(Succeed())
			managedResourceInstall2Secret.Name = managedResourceInstall2.Spec.SecretRefs[0].Name
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall2Secret), managedResourceInstall2Secret)).To(Succeed())
			Expect(managedResourceInstall2Secret.Immutable).To(Equal(pointer.Bool(true)))
			Expect(managedResourceInstall2Secret.Data).To(HaveLen(1))
		}

		It("Should successfully install gvisor to a single worker pool", func() {
			deployOnSingleWorkerPool()
		})

		It("Should successfully install gvisor to additional worker pool", func() {
			deployOnTwoWorkerPools()
		})

		It("Should successfully delete gvisor managed resources", func() {
			deployOnSingleWorkerPool()

			Expect(a.Delete(ctx, log, cr, cluster)).NotTo(HaveOccurred())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResource), managedResource)).To(BeNotFoundError())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceSecret), managedResourceSecret)).To(BeNotFoundError())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall), managedResourceInstall)).To(BeNotFoundError())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstallSecret), managedResourceInstallSecret)).To(BeNotFoundError())
		})

		It("Should successfully delete only one of the gvisor installations", func() {
			deployOnTwoWorkerPools()

			Expect(a.Delete(ctx, log, cr, cluster)).NotTo(HaveOccurred())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResource), managedResource)).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceSecret), managedResourceSecret)).To(Succeed())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall), managedResourceInstall)).To(BeNotFoundError())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstallSecret), managedResourceInstallSecret)).To(BeNotFoundError())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall2), managedResourceInstall2)).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(managedResourceInstall2Secret), managedResourceInstall2Secret)).To(Succeed())
		})
	})
})
