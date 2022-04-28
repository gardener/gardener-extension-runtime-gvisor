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

package container_runtime

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/gardener/gardener/test/framework"
	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const gVisorContainerRuntimeName = "gvisor"

var gVisorTimeout = 30 * time.Minute

var _ = ginkgo.Describe("gVisor tests", func() {
	f := framework.NewShootFramework(nil)

	f.Beta().Serial().CIt("should add, remove and upgrade worker pool with gVisor", func(ctx context.Context) {
		ginkgo.By("test adding new worker pool with containerd and gVisor")
		shoot := f.Shoot

		if len(shoot.Spec.Provider.Workers) == 0 {
			ginkgo.Skip("at least one worker pool is required in the test shoot.")
		}

		testWorker := shoot.Spec.Provider.Workers[0].DeepCopy()
		machineImage := testWorker.Machine.Image

		cloudProfile, err := f.GetCloudProfile(ctx)
		g.Expect(err).ToNot(g.HaveOccurred())

		if !supportsGVisor(cloudProfile.Spec.MachineImages, machineImage) {
			ginkgo.Skip(fmt.Sprintf("Skipping test as gVisor is not support on OS %q, version: %q, according to cloudprofile %q", machineImage.Name, *machineImage.Version, cloudProfile.GetName()))
		}

		ginkgo.By(fmt.Sprintf("OS %q, version: %q supports gVisor container runtime according to cloudprofile %q", machineImage.Name, *machineImage.Version, cloudProfile.GetName()))

		testWorker = configureWorkerForTesting(testWorker, true)

		shoot.Spec.Provider.Workers = append(shoot.Spec.Provider.Workers, *testWorker)

		ginkgo.By("adding gVisor worker pool")

		defer func(ctx context.Context, workerPoolName string) {
			ginkgo.By("removing gVisor worker pool after test execution")
			removeWorkerPool(ctx, f, workerPoolName)
		}(ctx, testWorker.Name)

		err = f.UpdateShoot(ctx, func(s *gardencorev1beta1.Shoot) error {
			s.Spec.Provider.Workers = shoot.Spec.Provider.Workers
			return nil
		})
		framework.ExpectNoError(err)

		// get the nodes of the worker pool and check if the node
		// labels of the worker pool contain the expected gVisor label
		nodeList := getGVisorNodes(ctx, f, testWorker)

		// deploy root pod
		rootPodExecutor := framework.NewRootPodExecutor(f.Logger, f.ShootClient, &nodeList.Items[0].Name, "kube-system")

		// gVisor requires containerd, so check that first
		containerdServiceCommand := fmt.Sprintf("systemctl is-active %s", extensionsv1alpha1.CRINameContainerD)
		executeCommand(ctx, rootPodExecutor, containerdServiceCommand, "active")

		// check that the binaries are available
		checkRunscShimBinary := fmt.Sprintf("[ -f %s/%s ] && echo 'found' || echo 'Not found'", string(extensionsv1alpha1.ContainerDRuntimeContainersBinFolder), "containerd-shim-runsc-v1")
		executeCommand(ctx, rootPodExecutor, checkRunscShimBinary, "found")

		checkRunscBinary := fmt.Sprintf("[ -f %s/%s ] && echo 'found' || echo 'Not found'", string(extensionsv1alpha1.ContainerDRuntimeContainersBinFolder), "runsc")
		executeCommand(ctx, rootPodExecutor, checkRunscBinary, "found")

		// check that containerd config.toml is configured for gVisor
		checkConfigurationCommand := "cat /etc/containerd/config.toml | grep -c 'containerd.runtimes.runsc'"
		executeCommand(ctx, rootPodExecutor, checkConfigurationCommand, "1")

		// deploy pod using gVisor RuntimeClass
		gVisorPod, err := deployGVisorPod(ctx, f.ShootClient.Client())
		g.Expect(err).ToNot(g.HaveOccurred())

		defer func(ctx context.Context, pod *corev1.Pod) {
			ginkgo.By("removing gVisor pod after test execution")
			err := f.ShootClient.Client().Delete(ctx, pod)
			g.Expect(err).ToNot(g.HaveOccurred())
		}(ctx, gVisorPod)

		// wait for it to run - implicitly checks that the pod has been scheduled to a node with gVisor enabled (would not start otherwise)
		err = framework.WaitUntilPodIsRunning(ctx, f.Logger, gVisorPod.Name, gVisorPod.Namespace, f.ShootClient)
		g.Expect(err).ToNot(g.HaveOccurred())

		// check kernel startup logs
		reader, err := framework.NewPodExecutor(f.ShootClient).Execute(ctx, gVisorPod.Namespace, gVisorPod.Name, gVisorPod.Spec.Containers[0].Name, "dmesg | grep -i -c gVisor")
		g.Expect(err).ToNot(g.HaveOccurred())

		response, err := ioutil.ReadAll(reader)
		g.Expect(err).ToNot(g.HaveOccurred())
		g.Expect(response).ToNot(g.BeNil())
		g.Expect(string(response)).To(g.Equal(fmt.Sprintf("%s\n", "1")))

		ginkgo.By("test removal of gVisor from worker pool")
		// remove gVisor from the worker pool and wait for the Shoot to be successfully reconciled.
		// That implies that gVisor has been removed successfully.
		removeGVisorFromWorker(ctx, f, testWorker.Name)

		ginkgo.By("test upgrading containerd pool to use gVisor")
		addGVisorToWorker(ctx, f, testWorker.Name)
	}, gVisorTimeout)

})

func getGVisorNodes(ctx context.Context, f *framework.ShootFramework, worker *gardencorev1beta1.Worker) *corev1.NodeList {
	return getNodeListWithLabel(ctx, f, worker, fmt.Sprintf(extensionsv1alpha1.ContainerRuntimeNameWorkerLabel, gVisorContainerRuntimeName), "true")
}

func getNodeListWithLabel(ctx context.Context, f *framework.ShootFramework, worker *gardencorev1beta1.Worker, nodeLabelKey, nodeLabelValue string) *corev1.NodeList {
	nodeList, err := framework.GetAllNodesInWorkerPool(ctx, f.ShootClient, &worker.Name)
	framework.ExpectNoError(err)
	g.Expect(len(nodeList.Items)).To(g.Equal(int(worker.Minimum)))

	for _, node := range nodeList.Items {
		value, found := node.Labels[nodeLabelKey]
		g.Expect(found).To(g.BeTrue())
		g.Expect(value).To(g.Equal(nodeLabelValue))
	}
	return nodeList
}

// configureWorkerForTesting configures the worker pool with test specific configuration such as a unique name and the CRI settings
func configureWorkerForTesting(worker *gardencorev1beta1.Worker, useGVisor bool) *gardencorev1beta1.Worker {
	allowedCharacters := "0123456789abcdefghijklmnopqrstuvwxyz"
	id, err := utils.GenerateRandomStringFromCharset(3, allowedCharacters)
	framework.ExpectNoError(err)

	worker.Name = fmt.Sprintf("test-%s", id)
	worker.Maximum = 1
	worker.Minimum = 1
	worker.CRI = &gardencorev1beta1.CRI{
		Name: gardencorev1beta1.CRINameContainerD,
	}

	if useGVisor {
		addGVisor(worker)
	}
	return worker
}

func addGVisor(worker *gardencorev1beta1.Worker) {
	worker.CRI.ContainerRuntimes = []gardencorev1beta1.ContainerRuntime{
		{
			Type: gVisorContainerRuntimeName,
		},
	}
}

func removeGVisorFromWorker(ctx context.Context, f *framework.ShootFramework, workerPoolName string) {
	err := f.UpdateShoot(ctx, func(s *gardencorev1beta1.Shoot) error {
		var workers []gardencorev1beta1.Worker
		for _, worker := range s.Spec.Provider.Workers {
			if worker.Name == workerPoolName {
				worker.CRI.ContainerRuntimes = []gardencorev1beta1.ContainerRuntime{}
			}
			workers = append(workers, worker)
		}
		s.Spec.Provider.Workers = workers
		return nil
	})
	framework.ExpectNoError(err)
}

func removeWorkerPool(ctx context.Context, f *framework.ShootFramework, workerPoolName string) {
	err := f.UpdateShoot(ctx, func(s *gardencorev1beta1.Shoot) error {
		var workers []gardencorev1beta1.Worker
		for _, worker := range s.Spec.Provider.Workers {
			if worker.Name == workerPoolName {
				continue
			}
			workers = append(workers, worker)
		}
		s.Spec.Provider.Workers = workers
		return nil
	})
	framework.ExpectNoError(err)
}

func addGVisorToWorker(ctx context.Context, f *framework.ShootFramework, workerPoolName string) {
	err := f.UpdateShoot(ctx, func(s *gardencorev1beta1.Shoot) error {
		var workers []gardencorev1beta1.Worker
		for _, worker := range s.Spec.Provider.Workers {
			if worker.Name == workerPoolName {
				worker.CRI.ContainerRuntimes = []gardencorev1beta1.ContainerRuntime{
					{
						Type: gVisorContainerRuntimeName,
					},
				}
			}
			workers = append(workers, worker)
		}
		s.Spec.Provider.Workers = workers
		return nil
	})
	framework.ExpectNoError(err)
}

// deployGVisorPod deploys a pod using the gVisor RuntimeClass.
func deployGVisorPod(ctx context.Context, c client.Client) (*corev1.Pod, error) {
	gVisorRuntimeClass := gVisorContainerRuntimeName
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
					Image: "eu.gcr.io/gardener-project/3rd/busybox:1.29.2",
					Command: []string{
						"sleep",
						"10000000",
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
func executeCommand(ctx context.Context, rootPodExecutor framework.RootPodExecutor, command, expected string) {
	response, err := rootPodExecutor.Execute(ctx, command)
	framework.ExpectNoError(err)
	g.Expect(response).ToNot(g.BeNil())
	g.Expect(string(response)).To(g.Equal(fmt.Sprintf("%s\n", expected)))
}

// supportsGVisor checks whether the given workerImage supports gVisor as container runtime
func supportsGVisor(cloudProfileImages []gardencorev1beta1.MachineImage, workerImage *gardencorev1beta1.ShootMachineImage) bool {
	var (
		cloudProfileImage *gardencorev1beta1.MachineImage
		machineVersion    *gardencorev1beta1.MachineImageVersion
	)

	for _, current := range cloudProfileImages {
		if current.Name == workerImage.Name {
			cloudProfileImage = &current
			break
		}
	}

	if cloudProfileImage == nil {
		return false
	}

	for _, version := range cloudProfileImage.Versions {
		if version.Version == *workerImage.Version {
			machineVersion = &version
			break
		}
	}

	if machineVersion == nil {
		return false
	}

	for _, cri := range machineVersion.CRI {
		if cri.Name != gardencorev1beta1.CRINameContainerD {
			continue
		}

		for _, runtime := range cri.ContainerRuntimes {
			if runtime.Type == gvisor.Type {
				return true
			}
		}
	}

	return false
}
