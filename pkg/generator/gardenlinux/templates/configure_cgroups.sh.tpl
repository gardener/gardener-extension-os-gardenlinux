#!/bin/bash

set -Eeuo pipefail

CGROUP_CMDLINE="/etc/kernel/cmdline.d/80-cgroup.cfg"

source "$(dirname $0)/g_functions.sh"

# these are the cgroup versions we want to have configured and that are actually running on the system
desired_cgroup="{{.cgroupVersion}}"
current_cgroup=$(check_current_cgroup)

# check if the system needs to be reconfigured
if [[ "$desired_cgroup" == "${current_cgroup%% *}" ]]; then
    echo "system already running with cgroup $desired_cgroup - exiting"
    exit 0
fi

# reconfigure the bootloader to include unified hierarchy (or not)
if [[ "$desired_cgroup" == "v1" ]]; then
    echo "configuring system to use cgroup v1"
    cat << '__EOF' > "$CGROUP_CMDLINE"
# Disable cgroup v2 support
CMDLINE_LINUX="$CMDLINE_LINUX systemd.unified_cgroup_hierarchy=0"
__EOF

elif [[ "$desired_cgroup" == "v2" ]]; then
    echo "configuring system to use cgroup v2"
    cat << '__EOF' > "$CGROUP_CMDLINE"
# Enable cgroup v2 support
CMDLINE_LINUX="$CMDLINE_LINUX systemd.unified_cgroup_hierarchy=1"
__EOF

else
    echo "desired cgroup version $desired_cgroup cannot be enabled, leaving system with $current_cgroup"
    exit 1
fi

# update bootloader
/usr/local/sbin/update-syslinux

# trigger a reboot by placing a control file into /var/run which will be picked up by
# gardener-restart-system.service
echo "scheduling a reboot to activate cgroup $desired_cgroup"
mkdir -p $(dirname $RESTART_CONTROL_FILE)
touch $RESTART_CONTROL_FILE
