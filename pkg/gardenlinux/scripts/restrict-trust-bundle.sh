#!/usr/bin/env bash
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

# Restrict the node's CA trust store to the gardener-delivered bundles by deactivating all
# distro-shipped CAs in /etc/ca-certificates.conf and re-running update-ca-certificates with
# /var/lib/ca-certificates-local as the only additional source.
#
# Idempotent: a marker comment is appended once; subsequent runs exit early.
# Files affected: /etc/ca-certificates.conf (rebuilt on /etc wipe), /etc/ssl/certs (rebuilt
# on /etc wipe). /var/lib/ca-certificates-local survives both the wipe and reboots, so a
# single --fresh here is enough together with the regular updatecacerts.service.

set -o errexit
set -o nounset
set -o pipefail

conf=/etc/ca-certificates.conf
marker='# gardener: replace-distro-trust-bundle'

if grep -qsF "$marker" "$conf" 2>/dev/null; then
  exit 0
fi

# Negate every not-yet-negated, non-comment entry. distro certs are identified by
# their path inside the read-only /usr image; we leave the entries in place but
# deactivate them so update-ca-certificates skips them during --fresh.
sed -i -E 's|^([^!#].*)|!\1|' "$conf"
echo "$marker" >> "$conf"

/usr/sbin/update-ca-certificates --fresh --localcertsdir /var/lib/ca-certificates-local
