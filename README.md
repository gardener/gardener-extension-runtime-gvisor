# [Gardener Extension for the gVisor Container Runtime Sandbox](https://gardener.cloud)

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/gardener-extension-runtime-gvisor)](https://api.reuse.software/info/github.com/gardener/gardener-extension-runtime-gvisor)
[![CI Build status](https://concourse.ci.gardener.cloud/api/v1/teams/gardener/pipelines/gardener-extension-runtime-gvisor-master/jobs/master-head-update-job/badge)](https://concourse.ci.gardener.cloud/teams/gardener/pipelines/gardener-extension-runtime-gvisor-master/jobs/master-head-update-job)
[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extension-runtime-gvisor)](https://goreportcard.com/report/github.com/gardener/gardener-extension-runtime-gvisor)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener). However, the project has grown to a size where it is very hard to extend, maintain, and test. With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics. This way, we can keep Gardener core clean and independent.

---

## How to use this Controller

This controller operates on the [ContainerRuntime](https://github.com/gardener/gardener/blob/master/docs/extensions/resources/containerruntime.md) resource in the `extensions.gardener.cloud/v1alpha1 API` group.

It manages objects that are requesting (`.spec.type=gvisor`) to enable the gVisor container runtime sandbox for a worker pool of a shoot cluster.

The ContainerRuntime can be configured in the shoot manifest in `.spec.povider.workers[].cri.containerRuntimes` an example can be found [here](example/shoot.yaml):

```yaml
kind: Shoot
apiVersion: core.gardener.cloud/v1beta1
metadata:
  name: gvisor-shoot
  namespace: garden-local
spec:
    ...
  provider:
    workers:
      - name: worker-xyz
    ...
        cri:
          name: containerd
          containerRuntimes:
            - type: gvisor
    ...
```

GVisor can be configured with additional configuration flags by adding them to the `configFlags` field in the providerConfig. Right now we only allow the `"net-raw"` flag to be set. All other flags are ignored.

```yaml
...
            - type: gvisor
              providerConfig:
                apiVersion: gvisor.os.extensions.gardener.cloud/v1alpha1
                kind: GVisorConfiguration
                configFlags:
                  "net-raw" "true"
                  ...
...
```

Based on the configuration in the shoot manifest the ContainerRuntime resource is created:

```yaml
---
apiVersion: extensions.gardener.cloud/v1alpha1
kind: ContainerRuntime
metadata:
  name: my-container-runtime
spec:
  binaryPath: /var/bin/containerruntimes
  type: gvisor
  providerConfig:
    apiVersion: gvisor.runtime.extensions.config.gardener.cloud/v1alpha1
    configFlags:
      net-raw: "true"
    kind: GVisorConfiguration
  workerPool:
    name: worker-ubuntu
    selector:
      matchLabels:
        worker.gardener.cloud/pool: worker-xyz
```

---

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start`. Please make sure to have the kubeconfig to the cluster you want to connect to ready in the `./dev/kubeconfig` file.

Static code checks and tests can be executed by running `make verify`. We are using Go modules for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extension-runtime-gvisor/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

- [Our landing page gardener.cloud](https://gardener.cloud/)
- ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
- ["Gardener Project Update" blog on kubernetes.io](https://kubernetes.io/blog/2019/12/02/gardener-project-update/)
- [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
- [GEP-10 (Additional Container Runtimes)](https://github.com/gardener/gardener/blob/master/docs/proposals/10-shoot-additional-container-runtimes.md)
- [Extensibility API documentation](https://github.com/gardener/gardener/tree/master/docs/extensions)
