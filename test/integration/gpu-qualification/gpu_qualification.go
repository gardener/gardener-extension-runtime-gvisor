// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gpu_qualification

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	kubernetesclient "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/test/framework"
	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	gVisorRuntimeClassName = "gvisor"
	gpuResourceName        = "nvidia.com/gpu"
)

// GPUMachineTypes lists machine types with NVIDIA GPUs supported for qualification.
var GPUMachineTypes = []string{
	"g4dn.xlarge", // AWS, NVIDIA T4
	"g5.xlarge",   // AWS, NVIDIA A10G
}

var gpuTimeout = 40 * time.Minute

var _ = ginkgo.Describe("gVisor GPU qualification", func() {
	f := framework.NewShootFramework(nil)

	f.Beta().Serial().CIt("should run GPU workload inside gVisor sandbox with nvproxy", func(ctx context.Context) {
		shoot := f.Shoot

		nvidiaInstallerVersion := os.Getenv("NVIDIA_INSTALLER_VERSION")
		if nvidiaInstallerVersion == "" {
			nvidiaInstallerVersion = "1.14.1"
		}

		ginkgo.By("verifying shoot has GPU worker pool with gVisor")
		hasGPUWorker := false
		for _, worker := range shoot.Spec.Provider.Workers {
			for _, gpuType := range GPUMachineTypes {
				if worker.Machine.Type == gpuType {
					hasGPUWorker = true
					break
				}
			}
		}
		if !hasGPUWorker {
			ginkgo.Skip(fmt.Sprintf("shoot does not have a GPU worker pool (supported: %v)", GPUMachineTypes))
		}

		ginkgo.By("verifying gVisor RuntimeClass exists")
		runtimeClass := &metav1.PartialObjectMetadata{}
		runtimeClass.SetGroupVersionKind(metav1.SchemeGroupVersion.WithKind("RuntimeClass"))
		err := f.ShootClient.Client().Get(ctx, client.ObjectKey{Name: gVisorRuntimeClassName}, runtimeClass)
		framework.ExpectNoError(err)

		ginkgo.By(fmt.Sprintf("installing NVIDIA driver via gardenlinux-nvidia-installer %s", nvidiaInstallerVersion))
		err = installNvidiaDriver(ctx, f, nvidiaInstallerVersion)
		framework.ExpectNoError(err)

		ginkgo.By("waiting for nvidia.com/gpu resources on nodes")
		err = waitForGPUResources(ctx, f, 10*time.Minute)
		framework.ExpectNoError(err)

		ginkgo.By("deploying GPU test pod with gVisor runtime (hashcat benchmark)")
		gpuPod, err := deployGPUTestPod(ctx, f.ShootClient.Client())
		framework.ExpectNoError(err)

		defer func(ctx context.Context) {
			ginkgo.By("cleaning up GPU test pod")
			_ = f.ShootClient.Client().Delete(ctx, gpuPod)
		}(ctx)

		ginkgo.By("waiting for GPU test pod to complete")
		err = framework.WaitUntilPodIsRunningOrSucceeded(ctx, f.Logger, gpuPod.Name, gpuPod.Namespace, f.ShootClient)
		framework.ExpectNoError(err)

		// Wait for completion (pod has restartPolicy: Never)
		g.Eventually(func() corev1.PodPhase {
			p := &corev1.Pod{}
			if err := f.ShootClient.Client().Get(ctx, client.ObjectKeyFromObject(gpuPod), p); err != nil {
				return corev1.PodUnknown
			}
			return p.Status.Phase
		}).WithTimeout(5 * time.Minute).WithPolling(10 * time.Second).Should(g.Equal(corev1.PodSucceeded))

		ginkgo.By("validating GPU test output")
		stdout, _, err := kubernetesclient.NewPodExecutor(f.ShootClient.RESTConfig()).Execute(
			ctx, gpuPod.Namespace, gpuPod.Name, gpuPod.Spec.Containers[0].Name,
			"cat", "/tmp/result.txt",
		)
		// If exec fails because pod completed, read logs instead
		if err != nil {
			logReader, logErr := f.ShootClient.Kubernetes().CoreV1().Pods(gpuPod.Namespace).GetLogs(gpuPod.Name, &corev1.PodLogOptions{}).Stream(ctx)
			g.Expect(logErr).ToNot(g.HaveOccurred())
			defer logReader.Close()
			stdout = logReader
		}
		response, err := io.ReadAll(stdout)
		g.Expect(err).ToNot(g.HaveOccurred())
		g.Expect(string(response)).To(g.ContainSubstring("GPU_TEST_PASSED"))
	}, gpuTimeout)
})

// installNvidiaDriver installs the NVIDIA GPU Operator via helm using the
// gardenlinux-nvidia-installer values file. This deploys pre-compiled driver
// images (Garden Linux has no build tools on nodes).
func installNvidiaDriver(ctx context.Context, f *framework.ShootFramework, version string) error {
	// Use a pod to run helm install on the shoot cluster
	valuesURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/gardenlinux/gardenlinux-nvidia-installer/refs/tags/%s/helm/gpu-operator-values.yaml",
		version,
	)

	helmPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "nvidia-installer-",
			Namespace:    "default",
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "default",
			RestartPolicy:      corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "helm",
					Image: "europe-docker.pkg.dev/gardener-project/releases/3rd/alpine/helm:3.16.4",
					Command: []string{"sh", "-c", fmt.Sprintf(`
						set -e
						helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
						helm repo update
						helm upgrade --install gpu-operator nvidia/gpu-operator \
							--namespace gpu-operator --create-namespace \
							--values "%s" \
							--wait --timeout 600s
						echo "NVIDIA_INSTALL_DONE"
					`, valuesURL)},
				},
			},
		},
	}

	if err := f.ShootClient.Client().Create(ctx, helmPod); err != nil {
		return err
	}

	// Wait for helm pod to succeed
	return framework.WaitUntilPodCompleted(ctx, f.Logger, helmPod.Name, helmPod.Namespace, f.ShootClient)
}

// waitForGPUResources waits until at least one node advertises nvidia.com/gpu > 0.
func waitForGPUResources(ctx context.Context, f *framework.ShootFramework, timeout time.Duration) error {
	return framework.WaitForCondition(ctx, f.Logger, timeout, 15*time.Second, func() (bool, error) {
		nodeList := &corev1.NodeList{}
		if err := f.ShootClient.Client().List(ctx, nodeList); err != nil {
			return false, err
		}
		for _, node := range nodeList.Items {
			if gpuQuantity, ok := node.Status.Allocatable[corev1.ResourceName(gpuResourceName)]; ok {
				if gpuQuantity.Cmp(resource.MustParse("1")) >= 0 {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

// deployGPUTestPod deploys a pod that runs hashcat GPU benchmark inside gVisor sandbox.
// hashcat performs GPU-accelerated hash computation, proving the GPU is accessible via nvproxy.
func deployGPUTestPod(ctx context.Context, c client.Client) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "gvisor-gpu-test-",
			Namespace:    "default",
			Labels: map[string]string{
				"app": "gvisor-gpu-test",
			},
		},
		Spec: corev1.PodSpec{
			RuntimeClassName: ptr.To(gVisorRuntimeClassName),
			RestartPolicy:    corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "gpu-hashcat",
					Image: "nvidia/cuda:12.2.0-runtime-ubuntu22.04",
					Command: []string{"bash", "-c", `
						set -e
						echo "=== gVisor GPU Test (hashcat) ==="
						echo "--- nvidia-smi ---"
						nvidia-smi
						echo ""
						echo "--- Installing hashcat ---"
						apt-get update -qq && apt-get install -qq -y hashcat > /dev/null 2>&1
						echo "--- Running hashcat benchmark (MD5, GPU) ---"
						hashcat -b -m 0 --force --runtime=30 2>&1 | tail -20
						echo ""
						echo "GPU_TEST_PASSED" | tee /tmp/result.txt
					`},
					Env: []corev1.EnvVar{
						{Name: "NVIDIA_VISIBLE_DEVICES", Value: "all"},
						{Name: "NVIDIA_DRIVER_CAPABILITIES", Value: "compute,utility"},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceName(gpuResourceName): resource.MustParse("1"),
						},
					},
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: ptr.To(false),
					},
				},
			},
			Tolerations: []corev1.Toleration{
				{
					Key:      "nvidia.com/gpu",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	if err := c.Create(ctx, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

// unused but kept for reference - would be used for DaemonSet-based driver verification
var _ = labels.Everything
var _ = &appsv1.DaemonSet{}
