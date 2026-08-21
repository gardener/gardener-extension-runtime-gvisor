// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gpu_qualification

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	kubernetesclient "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/kubernetes/health"
	"github.com/gardener/gardener/pkg/utils/retry"
	"github.com/gardener/gardener/test/framework"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-runtime-gvisor/test/integration/container-runtime/common"
)

const (
	defaultNvidiaInstallerVersion = "1.16.0"
	// alpineHelmImage is the Alpine-based Helm image used to run the helm
	// installer Pod that deploys the NVIDIA gpu-operator into the shoot cluster.
	// To mirror/update the image in the repository, you can use:
	// `docker buildx imagetools create --tag europe-docker.pkg.dev/gardener-project/releases/3rd/alpine-helm:3.21.4 alpine/helm:3.21.4`
	alpineHelmImage = "europe-docker.pkg.dev/gardener-project/releases/3rd/alpine/helm:3.21.4"
	gpuHashCatImage = "nvidia/cuda:12.2.0-runtime-ubuntu22.04"

	gVisorRuntimeClassName = "gvisor"
	gpuResourceName        = "nvidia.com/gpu"
	defaultPollInterval    = 5 * time.Second
	defaultTimeout         = 5 * time.Minute
)

// GPUMachineTypes lists machine types with NVIDIA GPUs supported for qualification.
var GPUMachineTypes = []string{
	"g4dn.xlarge", // AWS, NVIDIA T4
	"g5.xlarge",   // AWS, NVIDIA A10G
}

var gpuTimeout = 40 * time.Minute

var _ = Describe("gVisor GPU qualification", func() {
	f := framework.NewShootFramework(nil)

	f.Beta().Serial().CIt("should run GPU workload inside gVisor sandbox with nvproxy", func(ctx context.Context) {
		shoot := f.Shoot

		msg, skip, err := common.SkipGVisor(ctx, f)
		Expect(err).ToNot(HaveOccurred())
		if skip {
			Skip(msg)
		}

		By(msg)

		nvidiaInstallerVersion := os.Getenv("NVIDIA_INSTALLER_VERSION")
		if nvidiaInstallerVersion == "" {
			nvidiaInstallerVersion = defaultNvidiaInstallerVersion
		}

		By("verifying shoot has GPU worker pool with gVisor")
		gpuWorkerPoolMachineType := os.Getenv("GPU_WORKER_POOL_MACHINE_TYPE")
		var gpuWorker *gardencorev1beta1.Worker
		if gpuWorkerPoolMachineType == "" {
			for _, worker := range shoot.Spec.Provider.Workers {
				if slices.Contains(GPUMachineTypes, worker.Machine.Type) {
					gpuWorker = &worker
					break
				}
			}
			if gpuWorker == nil {
				// if GPU_WORKER_POOL_MACHINE_TYPE is provided, a worker pool is created
				Skip(fmt.Sprintf("shoot does not have a GPU worker pool (supported: %v)", GPUMachineTypes))
			}
		} else {
			for _, worker := range shoot.Spec.Provider.Workers {
				if worker.Machine.Type == gpuWorkerPoolMachineType {
					gpuWorker = &worker
					break
				}
			}

			if gpuWorker == nil {
				cfg, err := common.NewTestWorker(true, true)
				Expect(err).ToNot(HaveOccurred())

				_, cleanup := common.AddTestWorkerPool(ctx, f, cfg)
				defer cleanup()
			} else {
				By("verifying pre-existing GPU worker pool has gVisor runtime with nvproxy")
				Expect(common.HasGVisorRuntime(gpuWorker, true)).To(BeTrue(),
					"pre-existing worker pool %q (machine type %q) must have the gVisor container runtime with nvproxy enabled for GPU qualification", gpuWorker.Name, gpuWorkerPoolMachineType)
			}
		}

		By("verifying gVisor RuntimeClass exists")
		runtimeClass := &v1.RuntimeClass{}
		err = f.ShootClient.Client().Get(ctx, client.ObjectKey{Name: gVisorRuntimeClassName}, runtimeClass)
		framework.ExpectNoError(err)

		By(fmt.Sprintf("installing NVIDIA driver via gardenlinux-nvidia-installer %s", nvidiaInstallerVersion))
		err = installNvidiaDriver(ctx, f, nvidiaInstallerVersion)
		framework.ExpectNoError(err)

		By("waiting for nvidia.com/gpu resources on nodes")
		err = waitForGPUResources(ctx, f, 60*time.Minute)
		framework.ExpectNoError(err)

		By("deploying GPU test pod with gVisor runtime (hashcat benchmark)")
		gpuPod, err := deployGPUTestPod(ctx, f.ShootClient.Client())
		framework.ExpectNoError(err)

		defer func(ctx context.Context) {
			By("cleaning up GPU test pod")
			_ = f.ShootClient.Client().Delete(ctx, gpuPod)
		}(ctx)

		By("waiting for GPU test pod to complete")
		err = waitUntilPodCompleted(ctx, f.Logger, gpuPod.Name, gpuPod.Namespace, f.ShootClient, defaultTimeout)
		framework.ExpectNoError(err)

		// check if succeeded. Poll with Eventually to tolerate watch/cache delays: the
		// PodCompleted condition can be observed before Status.Phase has propagated to
		// PodSucceeded in the client's cache.
		Eventually(ctx, func(g Gomega) corev1.PodPhase {
			p := &corev1.Pod{}
			g.Expect(f.ShootClient.Client().Get(ctx, client.ObjectKeyFromObject(gpuPod), p)).To(Succeed())
			return p.Status.Phase
		}).WithPolling(1*time.Second).WithTimeout(30*time.Second).Should(Equal(corev1.PodSucceeded), func() string {
			logs, logErr := fetchPodLogs(ctx, f.ShootClient, gpuPod.Name, gpuPod.Namespace)
			if logErr != nil {
				return fmt.Sprintf("pod %q did not reach phase %q; additionally, failed to get pod logs: %v", gpuPod.Namespace+"/"+gpuPod.Name, corev1.PodSucceeded, logErr)
			}
			return fmt.Sprintf("pod %q did not reach phase %q; logs:\n%s", gpuPod.Namespace+"/"+gpuPod.Name, corev1.PodSucceeded, logs)
		})

		By("validating GPU test output")
		stdout, _, err := kubernetesclient.NewPodExecutor(f.ShootClient.RESTConfig()).Execute(
			ctx, gpuPod.Namespace, gpuPod.Name, gpuPod.Spec.Containers[0].Name,
			"cat", "/tmp/result.txt",
		)
		// If exec fails because pod completed, read logs instead
		if err != nil {
			logReader, logErr := f.ShootClient.Kubernetes().CoreV1().Pods(gpuPod.Namespace).GetLogs(gpuPod.Name, &corev1.PodLogOptions{}).Stream(ctx)
			Expect(logErr).ToNot(HaveOccurred())
			defer func() { _ = logReader.Close() }()
			stdout = logReader
		}
		response, err := io.ReadAll(stdout)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(response)).To(ContainSubstring("GPU_TEST_PASSED"))
	}, gpuTimeout)
})

// waitUntilPodCompleted waits until the pod with <podName> is completed successfully.
// If the pod does not complete in time or does not succeed, the pod logs are fetched
// and joined into the returned error to aid debugging.
func waitUntilPodCompleted(ctx context.Context, log logr.Logger, name, namespace string, c kubernetesclient.Interface, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := retry.Until(timeoutCtx, defaultPollInterval, func(ctx context.Context) (done bool, err error) {
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
		podLog := log.WithValues("pod", client.ObjectKeyFromObject(pod))

		if err := c.Client().Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, pod); err != nil {
			return retry.SevereError(err)
		}

		if pod.Status.Phase == corev1.PodFailed {
			return retry.SevereError(fmt.Errorf(`pod "%s/%s" failed`, namespace, name))
		}

		if !health.IsPodCompleted(pod.Status.Conditions) {
			podLog.Info("Waiting for Pod to be completed")
			return retry.MinorError(fmt.Errorf(`pod "%s/%s" is not completed`, namespace, name))
		}

		podLog.Info("Pod is completed now")
		return retry.Ok()
	})

	if err != nil {
		logs, logErr := fetchPodLogs(ctx, c, name, namespace)
		if logErr != nil {
			return errors.Join(err, fmt.Errorf("additionally, failed to get logs of pod %q: %w", namespace+"/"+name, logErr))
		}
		return errors.Join(err, fmt.Errorf("logs of pod %q:\n%s", namespace+"/"+name, logs))
	}

	return nil
}

// fetchPodLogs streams and returns the logs of the given pod.
func fetchPodLogs(ctx context.Context, c kubernetesclient.Interface, name, namespace string) (string, error) {
	logReader, err := c.Kubernetes().CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{}).Stream(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = logReader.Close() }()

	logs, err := io.ReadAll(logReader)
	if err != nil {
		return "", err
	}
	return string(logs), nil
}

// installNvidiaDriver installs the NVIDIA GPU Operator via helm using the
// gardenlinux-nvidia-installer values file. This deploys pre-compiled driver
// images (Garden Linux has no build tools on nodes).
func installNvidiaDriver(ctx context.Context, f *framework.ShootFramework, version string) error {
	// Use a pod to run helm install on the shoot cluster
	valuesURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/gardenlinux/gardenlinux-nvidia-installer/refs/tags/%s/helm/gpu-operator-values.yaml",
		version,
	)

	// The helm install needs cluster-wide permissions to manage the gpu-operator
	// installation (create namespaces, CRDs, RBAC, etc.). The default service
	// account is not allowed to do so, so create a dedicated service account
	// bound to the cluster-admin ClusterRole.
	saName := "nvidia-installer"
	if err := ensureHelmServiceAccount(ctx, f.ShootClient.Client(), saName, "default"); err != nil {
		return err
	}

	helmPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "nvidia-installer-",
			Namespace:    "default",
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: saName,
			RestartPolicy:      corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "helm",
					Image: alpineHelmImage,
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
	return waitUntilPodCompleted(ctx, f.Logger, helmPod.Name, helmPod.Namespace, f.ShootClient, defaultTimeout)
}

// ensureHelmServiceAccount creates a ServiceAccount in the given namespace and
// binds it to the cluster-admin ClusterRole so that the helm pod can manage the
// cluster-wide resources required by the NVIDIA GPU Operator.
func ensureHelmServiceAccount(ctx context.Context, c client.Client, name, namespace string) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := c.Create(ctx, sa); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-cluster-admin",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      name,
				Namespace: namespace,
			},
		},
	}
	if err := c.Create(ctx, crb); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// waitForGPUResources waits until at least one node advertises nvidia.com/gpu > 0.
func waitForGPUResources(ctx context.Context, f *framework.ShootFramework, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return retry.Until(timeoutCtx, defaultPollInterval, func(ctx context.Context) (done bool, err error) {
		nodeList := &corev1.NodeList{}
		if err := f.ShootClient.Client().List(ctx, nodeList); err != nil {
			return retry.SevereError(err)
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
			RuntimeClassName: new(gVisorRuntimeClassName),
			RestartPolicy:    corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "gpu-hashcat",
					Image: gpuHashCatImage,
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
