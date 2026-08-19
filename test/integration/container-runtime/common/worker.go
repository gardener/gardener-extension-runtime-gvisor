// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/gardener/gardener/test/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	configv1alpha1 "github.com/gardener/gardener-extension-runtime-gvisor/pkg/apis/config/v1alpha1"
)

// SkipGVisor checks whether the shoot's first worker pool supports the gVisor container runtime
// according to the cloud profile. Returns a skip=true result with an explanatory message if gVisor
// is not supported, or a confirmation message if it is.
func SkipGVisor(ctx context.Context, f *framework.ShootFramework) (msg string, skip bool, err error) {
	shoot := f.Shoot

	if len(shoot.Spec.Provider.Workers) == 0 {
		msg = "at least one worker pool is required in the test shoot."
		skip = true
		return
	}

	machineImage := shoot.Spec.Provider.Workers[0].Machine.Image

	cloudProfile, err := f.GetCloudProfile(ctx)
	if err != nil {
		return
	}

	if !supportsGVisor(cloudProfile.Spec.MachineImages, machineImage) {
		msg = fmt.Sprintf("Skipping test as gVisor is not support on OS %q, version: %q, according to cloudprofile %q", machineImage.Name, *machineImage.Version, cloudProfile.GetName())
		skip = true
		return
	}

	msg = fmt.Sprintf("OS %q, version: %q supports gVisor container runtime according to cloudprofile %q", machineImage.Name, *machineImage.Version, cloudProfile.GetName())
	return
}

// NewTestWorker builds a WorkerConfig from the environment. UseGVisor toggles the gVisor
// container runtime and gpuMachine requires the GPU_WORKER_POOL_MACHINE_TYPE env variable to
// select a GPU machine type. The optional TEST_IMAGE_TAG and GVISOR_VERSION env variables must
// either both be set or both be empty; an error is returned otherwise.
func NewTestWorker(useGVisor, gpuMachine bool) (*WorkerConfig, error) {
	cfg := WorkerConfig{
		UseGVisor:             useGVisor,
		TestImageTag:          os.Getenv("TEST_IMAGE_TAG"),
		ExpectedGVisorVersion: os.Getenv("GVISOR_VERSION"),
	}
	if gpuMachine {
		machineType := os.Getenv("GPU_WORKER_POOL_MACHINE_TYPE")
		if machineType == "" {
			return nil, fmt.Errorf("GPU_WORKER_POOL_MACHINE_TYPE must be set")
		}
		cfg.GPUMachineType = machineType
	}
	if (len(cfg.TestImageTag) == 0) != (len(cfg.ExpectedGVisorVersion) == 0) {
		return nil, fmt.Errorf("either both `TEST_IMAGE_TAG` and `GVISOR_VERSION` or none must be set")
	}
	return &cfg, nil
}

// HasGVisorRuntime reports whether the worker pool has the gVisor container runtime configured
// in its containerd CRI. If requireNVProxy is true, it additionally requires the gVisor
// GVisorConfiguration provider config to enable the nvproxy config flag (needed for GPU support).
func HasGVisorRuntime(worker *gardencorev1beta1.Worker, requireNVProxy bool) bool {
	if worker.CRI == nil {
		return false
	}

	var gvisor *gardencorev1beta1.ContainerRuntime
	for i := range worker.CRI.ContainerRuntimes {
		if worker.CRI.ContainerRuntimes[i].Type == GVisorContainerRuntimeName {
			gvisor = &worker.CRI.ContainerRuntimes[i]
			break
		}
	}
	if gvisor == nil {
		return false
	}

	if !requireNVProxy {
		return true
	}

	if gvisor.ProviderConfig == nil {
		return false
	}

	var providerConfig struct {
		ConfigFlags map[string]string `json:"configFlags"`
	}
	if err := json.Unmarshal(gvisor.ProviderConfig.Raw, &providerConfig); err != nil {
		return false
	}
	return providerConfig.ConfigFlags["nvproxy"] == "true"
}

// WorkerConfig holds the configuration used to derive a test worker pool.
type WorkerConfig struct {
	// UseGVisor indicates whether the gVisor container runtime should be added to the worker pool.
	UseGVisor bool
	// TestImageTag is the image tag of the gVisor test image to use (from the TEST_IMAGE_TAG env variable).
	TestImageTag string
	// ExpectedGVisorVersion is the gVisor version expected for the test image tag (from the GVISOR_VERSION env variable).
	ExpectedGVisorVersion string
	// GPUMachineType is the machine type to use for a GPU-enabled worker pool; empty means no GPU machine.
	GPUMachineType string
}

// ConfigureWorkerForTesting configures the worker pool with test specific configuration such as a unique name and the CRI settings
func (c WorkerConfig) ConfigureWorkerForTesting(f *framework.ShootFramework) *gardencorev1beta1.Worker {
	worker := f.Shoot.Spec.Provider.Workers[0].DeepCopy()

	allowedCharacters := "0123456789abcdefghijklmnopqrstuvwxyz"
	id, err := utils.GenerateRandomStringFromCharset(3, allowedCharacters)
	framework.ExpectNoError(err)

	worker.Name = fmt.Sprintf("test-%s", id)
	worker.Maximum = 1
	worker.Minimum = 1
	worker.CRI = &gardencorev1beta1.CRI{
		Name: gardencorev1beta1.CRINameContainerD,
	}

	if c.UseGVisor {
		c.addGVisor(f, worker, c.GPUMachineType != "")
	}
	if c.GPUMachineType != "" {
		worker.Machine.Type = c.GPUMachineType
	}

	return worker
}

// addGVisor adds the gVisor container runtime to the worker's containerd CRI. If withGPU is true,
// the gVisor config flags enable debug and nvproxy for GPU support. When a test image tag is set,
// a GVisorConfiguration provider config referencing that tag is attached to the container runtime.
func (c WorkerConfig) addGVisor(f *framework.ShootFramework, worker *gardencorev1beta1.Worker, withGPU bool) {
	var providerConfig *runtime.RawExtension

	if c.TestImageTag != "" {
		configFlags := map[string]string{}
		if withGPU {
			configFlags = map[string]string{"debug": "true", "nvproxy": "true"}
		}
		gvisorConfig := &configv1alpha1.GVisorConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: configv1alpha1.SchemeGroupVersion.String(),
				Kind:       "GVisorConfiguration",
			},
			TestImageTag: &c.TestImageTag,
			ConfigFlags:  &configFlags,
		}
		raw, err := json.Marshal(gvisorConfig)
		framework.ExpectNoError(err)
		providerConfig = &runtime.RawExtension{Raw: raw}
		f.Logger.Info("Using gVisorConfiguration for test worker pool", "imageTag", c.TestImageTag, "expectedGVisorVersion", c.ExpectedGVisorVersion)
	}

	worker.CRI.ContainerRuntimes = []gardencorev1beta1.ContainerRuntime{
		{
			Type:           GVisorContainerRuntimeName,
			ProviderConfig: providerConfig,
		},
	}
}
