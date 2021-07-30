# [Gardener Extension for Garden Linux OS](https://gardener.cloud)

[![CI Build status](https://concourse.ci.gardener.cloud/api/v1/teams/gardener/pipelines/gardener-extension-os-gardenlinux-master/jobs/master-head-update-job/badge)](https://concourse.ci.gardener.cloud/teams/gardener/pipelines/gardener-extension-os-gardenlinux-master/jobs/master-head-update-job)
[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extension-os-gardenlinux)](https://goreportcard.com/report/github.com/gardener/gardener-extension-os-gardenlinux)

This controller operates on the [`OperatingSystemConfig`](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md#cloud-config-user-data-for-bootstrapping-machines) resource in the `extensions.gardener.cloud/v1alpha1` API group. It manages those objects that are requesting [Garden Linux OS](https://gardenlinux.io/) configuration (`.spec.type=gardenlinux`):

```yaml
---
apiVersion: extensions.gardener.cloud/v1alpha1
kind: OperatingSystemConfig
metadata:
  name: pool-01-original
  namespace: default
spec:
  type: gardenlinux
  units:
    ...
  files:
    ...
```

Please find [a concrete example](example/40-operatingsystemconfig.yaml) in the `example` folder.

After reconciliation the resulting data will be stored in a secret within the same namespace (as the config itself might contain confidential data). The name of the secret will be written into the resource's `.status` field:

```yaml
...
status:
  ...
  cloudConfig:
    secretRef:
      name: osc-result-pool-01-original
      namespace: default
  command: /usr/bin/env bash <path>
  units:
  - docker-monitor.service
  - kubelet-monitor.service
  - kubelet.service
```

The secret has one data key `cloud_config` that stores the generation.

An example for a `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

This controller is based on revision [b5ba8164](https://github.com/gardener/gardener-extension-os-ubuntu-alicloud/commit/b5ba8164ed4c52872b1d4bd5ee3eaa4ed58da966) of [gardener-extension-os-ubuntu-alicloud](https://github.com/gardener/gardener-extension-os-ubuntu-alicloud). Its implementation is using the [`oscommon`](https://github.com/gardener/gardener/blob/master/extensions/pkg/controller/operatingsystemconfig/oscommon/README.md) library for operating system configuration controllers.

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

----

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start`. Please make sure to have the kubeconfig to the cluster you want to connect to ready in the `./dev/kubeconfig` file.
Static code checks and tests can be executed by running `make verify`. We are using Go modules for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extension-os-gardenlinux/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* ["Gardener Project Update" blog on kubernetes.io](https://kubernetes.io/blog/2019/12/02/gardener-project-update/)
* [Gardener Extensions Golang library](https://godoc.org/github.com/gardener/gardener/extensions/pkg)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
* [Extensibility API documentation](https://github.com/gardener/gardener/tree/master/docs/extensions)
