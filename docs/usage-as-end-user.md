# Using the Garden Linux extension with Gardener as end-user

The [`core.gardener.cloud/v1beta1.Shoot` resource](https://github.com/gardener/gardener/blob/master/example/90-shoot.yaml) declares a few fields that should be considered when this OS extension is used. It essentially allows you to configure [Garden Linux](https://github.com/gardenlinux/gardenlinux) specific settings from the `Shoot` manifest.

In this document we describe how this configuration looks like and under which circumstances your attention may be required.

## Declaring Garden Linux specific configuration

To configure Garden Linux specific settings, you can declare a `OperatingSystemConfiguration` in the `Shoot` manifest for each worker pool at `.spec.provider.workers[].machine.image.providerConfig`. 

An example `OperatingSystemConfiguration` would look like this:

```yaml
providerConfig:
  apiVersion: gardenlinux.os.extensions.gardener.cloud/v1alpha1
  kind: OperatingSystemConfiguration
  cgroupVersion: v2
  linuxSecurityModule: SELinux
```

Configuration of these settings is done by deploying configuration shell scripts and corresponding systemd units into Garden Linux and running them before the kubelet is started.

## Setting cgroup version of Garden Linux

Kubernetes version `>= v1.19` support the unified cgroup hierarchy (a.k.a. cgroup v2) on the worker nodes' operating system.

To configure cgroup v2, the following line can be included into the `OperatingSystemConfiguration`:

```yaml
  cgroupVersion: v2
```

If not specified, this setting will default to cgroup `v1`. Also, for Shoot clusters with K8S `< v1.19`, cgroup `v1` will be enforced. Changing this setting will trigger a reboot of the node during bootstrap. A reboot will not be performed if the kubelet is found to be running.

Setting the system to cgroup `v2` will reconfigure Garden Linux to have systemd use the unified cgroup hierarchy and will configure kubelet and containerd to use systemd as a cgroup driver.

### Possible values for `cgroupVersion` (case matters):

| value | result |
|---|---|
| `v1` | Garden Linux will be configured to use the classic cgroup hierarchy (cgroup v1) |
| `v2` | Garden Linux will be configured to use the unified cgroup hierarchy (cgroup v2) |

## Setting the Linux Security Module

This setting allows you to configure the Linux Security Module (lsm) to be `SELinux` or `AppArmor`. Certain Kubernetes workloads might require either lsm to be loaded at boot of the worker node and will fail to run if it is not active.

To configure SELinux, the following line can be included into the `OperatingSystemConfiguration`:

```yaml
  linuxSecurityModule: SELinux
```

If not specifief, this setting will default to `AppArmor`. Changing this setting will trigger a reboot of the node during bootstrap. A reboot will not be performed if the kubelet is found to be running.

### Possible values for `linuxSecurityModule` (case matters):

| value | result |
|---|---|
| `AppArmor` | Garden Linux will be configured with _AppArmor_ as lsm |
| `SELinux` | Garden Linux will be configured with _SELinux_ as lsm |
