#!/bin/bash

set -Eeuo pipefail

CONTAINERD_CONFIG="/etc/containerd/config.toml"

# checks if containerd has already running tasks to prevent touching a containerd with 
# already running containers
function check_no_running_containerd_tasks {
    containerd_runtime_status_dir=/run/containerd/io.containerd.runtime.v2.task/k8s.io

    # if the status dir for k8s.io namespace does not exist, there are no containers
    # in said namespace
    if [ ! -d $containerd_runtime_status_dir ]; then
        echo "$containerd_runtime_status_dir does not exists - no tasks in k8s.io namespace" 
        return 0
    fi

    # count the number of containerd tasks in the k8s.io namespace
    num_tasks=$(find /run/containerd/io.containerd.runtime.v2.task/k8s.io/ -maxdepth 1 -type d | wc -l)

    if [ "$num_tasks" -eq 0 ]; then
        echo "no active tasks in k8s.io namespace" 
        return 0
    fi

    echo "there are $num_tasks active tasks in the k8s.io containerd namespace - terminating"
    return 1
}

# reconfigures containerd to use systemd as a cgroup driver
function configure_containerd {
    echo "Configuring containerd cgroup driver to systemd"
    sed -i "s/SystemdCgroup *= *false/SystemdCgroup = true/g" "$CONTAINERD_CONFIG"
}

if check_no_running_containerd_tasks; then
    configure_containerd

    # in rare cases it could be that the kubelet.service was already running when
    # containerd got reconfigured so we restart it to force its ExecStartPre
    if systemctl is-active kubelet.service; then
        echo "triggering kubelet restart..."
        systemctl restart --no-block kubelet.service
    fi
fi
