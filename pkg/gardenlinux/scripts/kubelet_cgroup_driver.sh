#!/bin/bash

set -Eeuo pipefail

source "$(dirname $0)/g_functions.sh"

KUBELET_CONFIG="/var/lib/kubelet/config/kubelet"

# reconfigure the kubelet to use systemd as a cgroup driver on cgroup v2 enabled systems
function configure_kubelet {
    desired_cgroup_driver=$1

    if [ ! -s "$KUBELET_CONFIG" ]; then
        echo "$KUBELET_CONFIG does not exist" >&2
        return
    fi

    if [[ "$desired_cgroup_driver" == "systemd" ]]; then
        echo "Configuring kubelet to use systemd as cgroup driver"
        sed -i "s/cgroupDriver: cgroupfs/cgroupDriver: systemd/" "$KUBELET_CONFIG"
    else
        echo "Configuring kubelet to use cgroupfs as cgroup driver"
        sed -i "s/cgroupDriver: systemd/cgroupDriver: cgroupfs/" "$KUBELET_CONFIG"
    fi
}

# determine which cgroup driver the kubelet is currently configured with
function get_kubelet_cgroup_driver {
    kubelet_cgroup_driver=$(grep cgroupDriver "$KUBELET_CONFIG" | awk -F ':' '{print $2}' | sed 's/^\W//g')
    echo "$kubelet_cgroup_driver"
}

# determine which cgroup driver containerd is using - this requires that the SystemdCgroup is in containerds
# running config - if it has been removed from the configfile, this will fail
function get_containerd_cgroup_driver {
    systemd_cgroup_driver=$(containerd config dump | grep SystemdCgroup | awk -F '=' '{print $2}' | sed 's/^\W//g')

    if [ "$systemd_cgroup_driver"  == "true" ]; then
        echo systemd
        return
    else
        echo cgroupfs
        return
    fi
}

if [ "$(get_kubelet_cgroup_driver)" != "$(get_containerd_cgroup_driver)" ]; then
    configure_kubelet "$(get_containerd_cgroup_driver)"
else
    cgroup_driver=$(get_kubelet_cgroup_driver)
    echo "kubelet and containerd are configured with the same cgroup driver ($cgroup_driver) - nothing to do"
fi
