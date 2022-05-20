#!/bin/bash

set -Eeuo pipefail

LSM_CMDLINE="/etc/kernel/cmdline.d/90-lsm.cfg"
RESTART_CONTROL_FILE="/var/run/gardener/restart-required"

function check_current_lsm {
    if grep -q selinux /sys/kernel/security/lsm; then
        echo SELinux
    elif grep -q apparmor /sys/kernel/security/lsm; then
        echo AppArmor
    else
        echo none
    fi
}

desired_lsm={{.linuxSecurityModule}}
current_lsm=$(check_current_lsm)

if [[ "$desired_lsm" == "AppArmor" ]]; then
    echo "configuring system to use AppArmor as lsm"
    cat << '__EOF' > "$LSM_CMDLINE"
# Set AppArmor as Linux Security Module
CMDLINE_LINUX="$CMDLINE_LINUX security=apparmor"
__EOF

elif [[ "$desired_lsm" == "SELinux" ]]; then
    echo "configuring system to use SELinux as lsm"
    cat << '__EOF' > "$LSM_CMDLINE"
# Set SELinux as Linux Security Module
CMDLINE_LINUX="$CMDLINE_LINUX security=selinux"
__EOF

else
    echo "desired lsm $desired_lsm cannot be enabled, leaving system with $current_lsm"
    exit 1
fi

# update bootloader
/usr/local/sbin/update-syslinux

if [[ "$desired_lsm" == "${current_lsm}" ]]; then
    echo "system already running with $desired_lsm - not triggering a reboot"
    exit 0
else
    echo "scheduling a reboot to Linux security module $desired_lsm"
    mkdir -p $(dirname $RESTART_CONTROL_FILE)
    touch $RESTART_CONTROL_FILE
fi