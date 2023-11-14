#!/bin/bash

set -Eeuo pipefail

source "$(dirname $0)/g_functions.sh"

KUBELET_CONFIG="/var/lib/kubelet/config/kubelet"

# reconfigure the kubelet to use systemd as a cgroup driver
function configure_kubelet_cgroup_driver {
    if [ ! -s "$KUBELET_CONFIG" ]; then
        echo "$KUBELET_CONFIG does not exist" >&2
        return
    fi

    echo "Configuring kubelet cgroup driver to systemd"
    sed -i "s/cgroupDriver: cgroupfs/cgroupDriver: systemd/" "$KUBELET_CONFIG"
}

# do not change the kubelet's configuration on an existing system with running containers
if has_running_containerd_tasks; then
    echo "Skip configuring the kubelet cgroup driver on a node with running containers"
    return
fi

# all recent/supported Gardenlinux versions mount cgroupsV2 only.
# This extension version is only compatible with cgroupsv2-only GL versions, hence
# can only be used in Gardener installations that have no cgroupsV1 Gl versions configured.
configure_kubelet_cgroup_driver
