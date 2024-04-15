#!/bin/bash

set -Eeuo pipefail

KUBELET_CONFIG="/var/lib/kubelet/config/kubelet"

# reconfigure the kubelet to use systemd as a cgroup driver
echo "Configuring kubelet to use systemd as cgroup driver"
sed -i "s/cgroupDriver: cgroupfs/cgroupDriver: systemd/g" "$KUBELET_CONFIG"
