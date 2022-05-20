#!/bin/bash

set -Eeuo pipefail

function get_fs_of_directory {
    [ -z "$1" -o ! -d "$1" ] && return
    echo -n "$(stat -c %T -f "$1")"
}

function check_current_cgroup {
    # determining if the system is running cgroupv1 or cgroupv2
    # using systemd approach as in
    # https://github.com/systemd/systemd/blob/d6d450074ff7729d43476804e0e19c049c03141d/src/basic/cgroup-util.c#L2105-L2149

    CGROUP_ID="cgroupfs"
    CGROUP2_ID="cgroup2fs"
    TMPFS_ID="tmpfs"

    cgroup_dir_fs="$(get_fs_of_directory /sys/fs/cgroup)"

    if [[ "$cgroup_dir_fs" == "$CGROUP2_ID" ]]; then
        echo "v2"
        return
    elif [[ "$cgroup_dir_fs" == "$TMPFS_ID" ]]; then
        if [[ "$(get_fs_of_directory /sys/fs/cgroup/unified)" == "$CGROUP2_ID" ]]; then
            echo "v1 (cgroupv2systemd)"
            return
        fi
        if [[ "$(get_fs_of_directory /sys/fs/cgroup/systemd)" == "$CGROUP2_ID" ]]; then
            echo "v1 (cgroupv2systemd232)"
            return
        fi
        if [[ "$(get_fs_of_directory /sys/fs/cgroup/systemd)" == "$CGROUP_ID" ]]; then
            echo "v1"
            return
        fi
    fi
    # if we came this far despite all those returns, it means something went wrong
    echo "failed to determine cgroup version for this system" >&2
    exit 1
}

# reconfigures containerd to use systemd as a cgroup driver on cgroup v2 enabled systems
function configure_containerd {
    desired_cgroup=$1
    CONTAINERD_CONFIG="/etc/containerd/config.toml"

    if [ ! -e "$CONTAINERD_CONFIG" ]; then
        echo "$CONTAINERD_CONFIG does not exist" >&2
        return
    fi

    if [[ "$desired_cgroup" == "v2" ]]; then
        echo "Configuring containerd cgroup driver to systemd"
        sed -i "s/SystemdCgroup *= *false/SystemdCgroup = true/" "$CONTAINERD_CONFIG"
    else
        echo "Configuring containerd cgroup driver to cgroupfs"
        sed -i "s/SystemdCgroup *= *true/SystemdCgroup = false/" "$CONTAINERD_CONFIG"
    fi
}

configure_containerd "$(check_current_cgroup)"
