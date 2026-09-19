#!/usr/bin/env bash
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

# This script is installed at /var/lib/gardenlinux/etc-setup-hooks/ and is executed by
# GardenLinux after the /etc overlay is wiped during in-place OS version
# upgrade. It restores only the minimum /etc state needed to bring gardener-node-agent
# back up - exactly what gardener-node-init does on a fresh node - and then hands control
# back to gardener-node-agent, which re-applies the full OperatingSystemConfig itself.
#
# Everything gardener-node-agent needs to run (its binary in /opt/bin, its config file,
# kubeconfig, cluster CA, kubelet PKI, containerd image store) lives under /opt and /var/lib
# and survives the wipe, so no certificate re-issuance is needed. The /etc bits restored here:
#   * the OS system trust store (/etc/ssl/certs), rebuilt from the surviving /var CA sources so
#     gardener-node-agent can pull its own image from a CA-fronted/private registry;
#   * containerd config + ExecStart drop-in, so containerd can start (gardener-node-agent
#     re-mutates the config on its next reconcile);
#   * the gardener-node-agent systemd unit, so the agent can start.
#
# gardener-node-agent then re-applies the rest of the OperatingSystemConfig. To force it to
# recompute and re-apply all files/units onto the wiped /etc (instead of skipping work
# because its cached state predates the wipe), its two state files under /var/lib are
# removed at the end.

### -------IMPORTANT NOTE-------
# This script is supposed to mimic the behavior of gardener-node-init, which is run on a fresh node. It is therefore
# important that this script is kept in sync with gardener-node-init
# https://github.com/gardener/gardener/blob/release-v1.150/pkg/component/extensions/operatingsystemconfig/nodeinit/templates/scripts/init.tpl.sh

set -o errexit
set -o nounset
set -o pipefail

echo "> Restoring minimal /etc state after overlay wipe (init-equivalent)"

# Tee to /dev/console ifwritable so output is visible via the cloud provider's out-of-band serial console even when
# the node never rejoins the cluster and the journal is unreachable.
if [ -w /dev/console ]; then
  exec > >(tee /dev/console) 2>&1 || true
fi

fatal() {
  echo "FATAL: $*" >&2
  exit 1
}

# --- verify gardener-node-agent prerequisites survived the wipe ---
# The binary lives under /opt and the config under /var/lib, so both must still be present.
# If either is missing the unit would only crash-loop silently.
NODE_AGENT_BINARY="{{ .NodeAgentBinaryPath }}"
NODE_AGENT_CONFIG_DIR="{{ .NodeAgentConfigDir }}"

if [ ! -x "${NODE_AGENT_BINARY}" ]; then
  fatal "gardener-node-agent binary ${NODE_AGENT_BINARY} is missing or not executable; cannot start the agent after the /etc wipe"
fi

if ! ls "${NODE_AGENT_CONFIG_DIR}"/config-*.yaml >/dev/null 2>&1; then
  fatal "no gardener-node-agent config file found in ${NODE_AGENT_CONFIG_DIR}; cannot start the agent after the /etc wipe"
fi

# --- restore the OS system trust store (registry CA / Gardener CA bundle) ---
# The /etc wipe destroys the merged trust bundle under /etc/ssl/certs, so the registry CA (and
# the Gardener ROOTcerts bundle) are no longer trusted. If gardener-node-agent's own image must
# be pulled from a registry fronted by that CA (e.g. a private registry), the pull fails with an
# unknown-CA TLS error BEFORE gardener-node-agent can re-apply the OperatingSystemConfig that
# would rebuild trust - a chicken-and-egg deadlock. We therefore rebuild trust here, up front.
#
# Best-effort: must never abort the hook (which would leave gardener-node-agent un-started), so failures are only logged.
UPDATE_CA_SCRIPT="{{ .UpdateCACertificatesScriptPath }}"
restore_ca_trust() {
  if [ -x "${UPDATE_CA_SCRIPT}" ]; then
    "${UPDATE_CA_SCRIPT}"
  else
    echo "WARNING: ${UPDATE_CA_SCRIPT} not found; the OS system trust store cannot be rebuilt from surviving /var CA sources" >&2
    return 1
  fi
}

if ! restore_ca_trust; then
  echo "WARNING: failed rebuilding the OS system trust store; gardener-node-agent may be unable to pull images from a CA-fronted registry until it reconciles" >&2
fi

# --- containerd default config ---
# containerd's config is generated on the node (it is not part of the OperatingSystemConfig
# files). The default config plus the ExecStart drop-in below are enough for containerd -
# and therefore gardener-node-agent - to start; gardener-node-agent mutates the config
# further (cgroup driver, sandbox image, registry config path) on its next reconcile.
setup_containerd() {
  if [ ! -s /etc/containerd/config.toml ]; then
    mkdir -p /etc/containerd/
    containerd config default > /etc/containerd/config.toml
    chmod 0644 /etc/containerd/config.toml
  fi

  mkdir -p /etc/systemd/system/containerd.service.d
  cat <<EOF > /etc/systemd/system/containerd.service.d/11-exec_config.conf
[Service]
ExecStart=
ExecStart=/usr/bin/containerd --config=/etc/containerd/config.toml
EOF
  chmod 0644 /etc/systemd/system/containerd.service.d/11-exec_config.conf
}

if ! setup_containerd; then
  echo "WARNING: failed writing containerd config; continuing so gardener-node-agent is still started" >&2
fi

# --- gardener-node-agent systemd unit ---
# Mirrors pkg/nodeagent/bootstrap.Bootstrap: write the unit at 0644, then enable and start it.
# The unit name and content are taken from the OperatingSystemConfig / gardener core constants, so they stay in
# sync with gardener automatically.
mkdir -p /etc/systemd/system
cat << EOF | base64 -d > "/etc/systemd/system/{{ .NodeAgentUnitName }}"
{{ .NodeAgentUnitContentB64 }}
EOF
chmod 0644 /etc/systemd/system/{{ .NodeAgentUnitName }}

systemctl daemon-reload

# Bring containerd up first so gardener-node-agent can talk to it. Best-effort: if containerd
# does not come up immediately, gardener-node-agent retries, so do not abort the hook here.
if ! systemctl enable --now containerd.service; then
  echo "WARNING: containerd.service did not start cleanly; gardener-node-agent will retry once it is up" >&2
fi

# Drop gardener-node-agent's cached OSC state so it re-applies the full OperatingSystemConfig
# onto the wiped /etc.
rm -f {{ .LastComputedOSCChangesFilePath }}
rm -f {{ .LastAppliedOSCFilePath }}

# Enable and start gardener-node-agent. Use a blocking start (not --no-block) so an activation
# failure is surfaced as a non-zero hook exit instead of silently succeeding while the unit
# crash-loops.
systemctl enable {{ .NodeAgentUnitName }}
if ! systemctl restart {{ .NodeAgentUnitName }}; then
  systemctl status --no-pager {{ .NodeAgentUnitName }} >&2 || true
  fatal "failed to start {{ .NodeAgentUnitName }} after the /etc wipe"
fi

echo "> Done restoring minimal /etc state; gardener-node-agent will re-apply the OperatingSystemConfig"
