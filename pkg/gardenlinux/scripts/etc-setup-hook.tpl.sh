#!/usr/bin/env bash
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

# This script is installed at /var/lib/gardenlinux/etc-setup-hooks/ and is executed
# by GardenLinux via run-parts after the /etc overlay is wiped during a wipe-based
# in-place OS version upgrade. It restores all /etc state that Gardener manages so
# that the node rejoins the cluster without a full re-bootstrap.
#
# All credentials and data in /var/lib/ (GNA kubeconfig, kubelet certs, containerd
# image store) survive the wipe unchanged, so no certificate re-issuance is needed.

set -o errexit
set -o nounset
set -o pipefail

echo "> Restoring /etc state after overlay wipe"

# --- containerd ---
if [ ! -s /etc/containerd/config.toml ]; then
  mkdir -p /etc/containerd/
  containerd config default > /etc/containerd/config.toml
  chmod 0644 /etc/containerd/config.toml
fi

mkdir -p /etc/systemd/system/containerd.service.d
printf '%s' '[Service]
ExecStart=
ExecStart=/usr/bin/containerd --config=/etc/containerd/config.toml' \
  > /etc/systemd/system/containerd.service.d/11-exec_config.conf
chmod 0644 /etc/systemd/system/containerd.service.d/11-exec_config.conf

printf '%s' '[Service]
LimitMEMLOCK=67108864
LimitNOFILE=1048576' \
  > /etc/systemd/system/containerd.service.d/override.conf
chmod 0644 /etc/systemd/system/containerd.service.d/override.conf

# --- systemd units ---
# Unit content is base64-encoded at generation time to avoid any shell injection
# via heredoc delimiters or metacharacters in the content.
{{ range .Units -}}
mkdir -p /etc/systemd/system
printf '%s' '{{ .ContentB64 }}' | base64 -d > /etc/systemd/system/'{{ .Name }}'
{{ end }}
systemctl daemon-reload

# enable+start containerd first so other units that depend on it can start
systemctl enable containerd.service
systemctl restart containerd.service

# enable all units from the OSC; start only non-template units
{{ range .Units -}}
systemctl enable '{{ .Name }}'
systemctl restart --no-block '{{ .Name }}' || true
{{ end }}
echo "> Done restoring /etc state"

# Remove the last-applied OSC snapshot so GNA re-applies all files and units
# on the next reconcile. Without this, GNA's diff against the old snapshot would
# only re-write files that changed between OS versions — leaving unchanged files
# (journald.conf, CA bundle in /etc/pki, etc.) missing from the wiped /etc.
rm -f /var/lib/gardener-node-agent/last-applied-osc.yaml
