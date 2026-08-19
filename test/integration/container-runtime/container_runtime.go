// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package container_runtime

import (
	"context"
	"fmt"
	"io"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kubernetesclient "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/test/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-runtime-gvisor/test/integration/container-runtime/common"
)

var gVisorTimeout = 30 * time.Minute

var _ = Describe("gVisor tests", func() {
	f := framework.NewShootFramework(nil)

	f.Beta().Serial().CIt("should add, remove and upgrade worker pool with gVisor (CPU)", func(ctx context.Context) {
		By("test adding new worker pool with containerd and gVisor")

		msg, skip, err := common.SkipGVisor(ctx, f)
		Expect(err).ToNot(HaveOccurred())
		if skip {
			Skip(msg)
		}

		By(msg)

		cfg, err := common.NewTestWorker(true, false)
		Expect(err).ToNot(HaveOccurred())

		By("adding gVisor worker pool")

		testWorker, cleanup := common.AddTestWorkerPool(ctx, f, cfg)
		defer cleanup()

		// get the nodes of the worker pool and check if the node
		// labels of the worker pool contain the expected gVisor label
		nodeList, err := common.GetGVisorNodes(ctx, f, testWorker)
		Expect(err).ToNot(HaveOccurred())

		By("deploy root pod")
		rootPodExecutor := framework.NewRootPodExecutor(f.Logger, f.ShootClient, &nodeList.Items[0].Name, "kube-system")

		// gVisor requires containerd, so check that first
		containerdServiceCommand := []string{"systemctl", "is-active", "containerd"}
		executeCommand(ctx, rootPodExecutor, containerdServiceCommand, "active")

		// check that the binaries are available
		checkRunscShimBinary := []string{"sh", "-c", fmt.Sprintf("[ -f %s/%s ] && echo 'found' || echo 'Not found'", extensionsv1alpha1.ContainerDRuntimeContainersBinFolder, "containerd-shim-runsc-v1")}
		executeCommand(ctx, rootPodExecutor, checkRunscShimBinary, "found")

		checkRunscBinary := []string{"sh", "-c", fmt.Sprintf("[ -f %s/%s ] && echo 'found' || echo 'Not found'", extensionsv1alpha1.ContainerDRuntimeContainersBinFolder, "runsc")}
		executeCommand(ctx, rootPodExecutor, checkRunscBinary, "found")

		// check expected gVisor version
		if cfg.ExpectedGVisorVersion != "" {
			expectedOutput := fmt.Sprintf("runsc version release-%s", cfg.ExpectedGVisorVersion)
			checkRunscBinaryVersion := []string{"sh", "-c", fmt.Sprintf("%s/%s --version | grep version", extensionsv1alpha1.ContainerDRuntimeContainersBinFolder, "runsc")}
			executeCommand(ctx, rootPodExecutor, checkRunscBinaryVersion, expectedOutput)
		}

		// check that containerd config.toml is configured for gVisor
		checkConfigurationCommand := []string{"sh", "-c", "cat /etc/containerd/config.toml | grep -c 'containerd.runtimes.runsc'"}
		executeCommand(ctx, rootPodExecutor, checkConfigurationCommand, "2")

		// deploy pod using gVisor RuntimeClass
		gVisorPod, err := deployGVisorPod(ctx, f.ShootClient.Client())
		Expect(err).ToNot(HaveOccurred())

		defer func(ctx context.Context, pod *corev1.Pod) {
			By("removing gVisor pod after test execution")
			err := f.ShootClient.Client().Delete(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}(ctx, gVisorPod)

		// wait for it to run - implicitly checks that the pod has been scheduled to a node with gVisor enabled (would not start otherwise)
		err = framework.WaitUntilPodIsRunning(ctx, f.Logger, gVisorPod.Name, gVisorPod.Namespace, f.ShootClient)
		Expect(err).ToNot(HaveOccurred())

		// check kernel startup logs
		stdout, _, err := kubernetesclient.NewPodExecutor(f.ShootClient.RESTConfig()).Execute(ctx, gVisorPod.Namespace, gVisorPod.Name, gVisorPod.Spec.Containers[0].Name, "sh", "-c", "dmesg | grep -i -c gVisor")
		Expect(err).ToNot(HaveOccurred())
		response, err := io.ReadAll(stdout)
		Expect(err).ToNot(HaveOccurred())
		Expect(response).ToNot(BeNil())
		Expect(string(response)).To(Equal(fmt.Sprintf("%s\n", "1")))

		By("test removal of gVisor from worker pool")
		// remove gVisor from the worker pool and wait for the Shoot to be successfully reconciled.
		// That implies that gVisor has been removed successfully.
		common.SetWorkerContainerRuntimes(ctx, f, testWorker.Name, nil)

		By("test upgrading containerd pool to use gVisor")
		common.SetWorkerContainerRuntimes(ctx, f, testWorker.Name, []gardencorev1beta1.ContainerRuntime{
			{Type: common.GVisorContainerRuntimeName},
		})
	}, gVisorTimeout)

})

// deployGVisorPod deploys a pod using the gVisor RuntimeClass.
func deployGVisorPod(ctx context.Context, c client.Client) (*corev1.Pod, error) {
	gVisorRuntimeClass := common.GVisorContainerRuntimeName
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "gvisor",
			Namespace:    "default",
		},
		Spec: corev1.PodSpec{
			RuntimeClassName: &gVisorRuntimeClass,
			Containers: []corev1.Container{
				{
					Name:  "gvisor-container",
					Image: "europe-docker.pkg.dev/gardener-project/releases/3rd/busybox:1.29.3",
					Command: []string{
						"sleep",
						"10000000",
					},
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: new(bool),
					},
				},
			},
		},
	}
	if err := c.Create(ctx, &pod); err != nil {
		return nil, err
	}
	return &pod, nil
}

// executeCommand executes a command on the host and checks the returned result
func executeCommand(ctx context.Context, rootPodExecutor framework.RootPodExecutor, command []string, expected string) {
	response, err := rootPodExecutor.Execute(ctx, command...)
	framework.ExpectNoError(err)
	Expect(response).ToNot(BeNil())
	Expect(string(response)).To(Equal(fmt.Sprintf("%s\n", expected)))
}
