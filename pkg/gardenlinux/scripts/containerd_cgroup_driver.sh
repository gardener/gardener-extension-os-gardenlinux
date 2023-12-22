#!/bin/bash

set -Eeuo pipefail

source "$(dirname $0)/g_functions.sh"

# reconfigures containerd to use systemd as a cgroup driver
function configure_containerd {
    CONTAINERD_CONFIG="/etc/containerd/config.toml"

    if [ ! -s "$CONTAINERD_CONFIG" ]; then
        echo "$CONTAINERD_CONFIG does not exist" >&2
        return
    fi

    echo "Setting containerd cgroup driver to systemd"
    sed -i "s/SystemdCgroup *= *false/SystemdCgroup = true/" "$CONTAINERD_CONFIG"
}

# do not change containerd's configuration on an existing system with running containers
if has_running_containerd_tasks; then
    echo "Skip configuring the containerd cgroup driver on a node with running containers"
    exit 0
fi

# all recent/supported Gardenlinux versions mount cgroupsV2 only.
# This extension version is only compatible with cgroupsv2-only GL versions, hence
# can only be used in Gardener installations that have no cgroupsV1 Gl versions configured.
configure_containerd
