#!/bin/bash

set -Eeuo pipefail

function has_running_containerd_tasks {
    containerd_runtime_status_dir=/run/containerd/io.containerd.runtime.v2.task/k8s.io

    # if the status dir for k8s.io namespace does not exist, there are no containers
    # in said namespace
    if [ ! -d $containerd_runtime_status_dir ]; then
        echo "$containerd_runtime_status_dir does not exists - no tasks in k8s.io namespace" 
        return 1
    fi

    # count the number of containerd tasks in the k8s.io namespace
    num_tasks=$(ls -1 /run/containerd/io.containerd.runtime.v2.task/k8s.io/ | wc -l)

    if [ "$num_tasks" -eq 0 ]; then
        echo "no active tasks in k8s.io namespace" 
        return 1
    fi

    echo "there are $num_tasks active tasks in the k8s.io containerd namespace"
    return 0
}
