#!/bin/bash

set -Eeuo pipefail

source "$(dirname $0)/g_functions.sh"

# reconfigure the kubelet to use systemd as a cgroup driver on cgroup v2 enabled systems
function configure_kubelet {
    desired_cgroup=$1
    KUBELET_CONFIG="/var/lib/kubelet/config/kubelet"

    if [ ! -s "$KUBELET_CONFIG" ]; then
        echo "$KUBELET_CONFIG does not exist" >&2
        return
    fi

    if [[ "$desired_cgroup" == "v2" ]]; then
        echo "Configuring kubelet to use systemd as cgroup driver"
        sed -i "s/cgroupDriver: cgroupfs/cgroupDriver: systemd/" "$KUBELET_CONFIG"
    else
        echo "Configuring kubelet to use cgroupfs as cgroup driver"
        sed -i "s/cgroupDriver: systemd/cgroupDriver: cgroupfs/" "$KUBELET_CONFIG"
    fi
}

if check_running_containerd_tasks; then
    configure_kubelet "$(check_current_cgroup)"
fi
