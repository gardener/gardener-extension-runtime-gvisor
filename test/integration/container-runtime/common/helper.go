// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/test/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

// GVisorContainerRuntimeName is the container runtime type name used to identify gVisor nodes and worker pools.
const GVisorContainerRuntimeName = "gvisor"

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

// GetGVisorNodes returns all nodes in the given worker pool that have the gVisor container runtime label set.
func GetGVisorNodes(ctx context.Context, f *framework.ShootFramework, worker *gardencorev1beta1.Worker) (*corev1.NodeList, error) {
	return getNodeListWithLabel(ctx, f, worker, fmt.Sprintf(extensionsv1alpha1.ContainerRuntimeNameWorkerLabel, GVisorContainerRuntimeName), "true")
}

func getNodeListWithLabel(ctx context.Context, f *framework.ShootFramework, worker *gardencorev1beta1.Worker, nodeLabelKey, nodeLabelValue string) (*corev1.NodeList, error) {
	nodeList, err := framework.GetAllNodesInWorkerPool(ctx, f.ShootClient, &worker.Name)
	if err != nil {
		return nil, err
	}
	if len(nodeList.Items) < int(worker.Minimum) {
		return nil, fmt.Errorf("worker %s does not have enough nodes", worker.Name)
	}

	for _, node := range nodeList.Items {
		value, found := node.Labels[nodeLabelKey]
		if !found {
			return nil, fmt.Errorf("node %s does not have label %s", node.Name, nodeLabelKey)
		}
		if value != nodeLabelValue {
			return nil, fmt.Errorf("node %s has label %s instead of %s", node.Name, nodeLabelKey, nodeLabelValue)
		}
	}
	return nodeList, nil
}

// AddWorkerPool appends the given worker pool to the shoot spec and waits for the shoot to be reconciled.
func AddWorkerPool(ctx context.Context, f *framework.ShootFramework, worker *gardencorev1beta1.Worker) error {
	f.Shoot.Spec.Provider.Workers = append(f.Shoot.Spec.Provider.Workers, *worker)
	return f.UpdateShoot(ctx, func(s *gardencorev1beta1.Shoot) error {
		s.Spec.Provider.Workers = f.Shoot.Spec.Provider.Workers
		return nil
	})
}

// AddTestWorkerPool derives a test worker pool from the given config, adds it to the shoot, and
// returns a cleanup function that removes it again. The caller should `defer` the returned function
// so the pool is removed even if the test fails after this call.
func AddTestWorkerPool(ctx context.Context, f *framework.ShootFramework, cfg *WorkerConfig) (worker *gardencorev1beta1.Worker, cleanup func()) {
	testWorker := cfg.ConfigureWorkerForTesting(f)

	cleanup = func() {
		By("removing gVisor worker pool after test execution")
		RemoveWorkerPool(ctx, f, testWorker.Name)
	}

	Expect(AddWorkerPool(ctx, f, testWorker)).ToNot(HaveOccurred())
	return testWorker, cleanup
}

// RemoveWorkerPool removes the worker pool with the given name from the shoot spec and waits for reconciliation.
func RemoveWorkerPool(ctx context.Context, f *framework.ShootFramework, workerPoolName string) {
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

// SetWorkerContainerRuntimes replaces the container runtimes of the worker pool with the given name
// and waits for the shoot to be reconciled. Passing an empty slice removes all container runtimes
// (e.g. gVisor) from the pool.
func SetWorkerContainerRuntimes(ctx context.Context, f *framework.ShootFramework, workerPoolName string, containerRuntimes []gardencorev1beta1.ContainerRuntime) {
	err := f.UpdateShoot(ctx, func(s *gardencorev1beta1.Shoot) error {
		for i := range s.Spec.Provider.Workers {
			if s.Spec.Provider.Workers[i].Name == workerPoolName {
				s.Spec.Provider.Workers[i].CRI.ContainerRuntimes = containerRuntimes
			}
		}
		return nil
	})
	framework.ExpectNoError(err)
}
