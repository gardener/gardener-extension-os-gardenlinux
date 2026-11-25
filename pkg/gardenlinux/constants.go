// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package gardenlinux

import (
	"embed"
)

var (
	//go:embed scripts/*
	Templates embed.FS
)

const (
	// ScriptLocation is the location that Gardener configuration scripts end up on Garden Linux
	ScriptLocation = "/opt/gardener/bin"

	// ScriptPermissions is the file permission used for Gardener configuration scripts installed on Garden Linux.
	ScriptPermissions = int32(0755)

	// PathEtcSetupHook is the path of the etc-setup hook script on the worker node.
	// GardenLinux runs every executable file in this directory (gated by the executable bit) after the
	// /etc overlay is wiped during an in-place OS version upgrade, so the script
	// must be installed with executable permissions.
	PathEtcSetupHook = "/var/lib/gardenlinux/etc-setup-hooks/00-gardener"

	// PathRestrictTrustBundle is the path of the script that restricts the node's CA trust store to
	// the gardener-delivered bundles by deactivating distro-shipped CAs. Rendered and delivered when
	// the operating-system config requests a strict (replace-distro) trust bundle.
	PathRestrictTrustBundle = ScriptLocation + "/restrict-trust-bundle.sh"

	// OSTypeGardenLinux is a constant for the Garden Linux extension OS type.
	OSTypeGardenLinux = "gardenlinux"

	// OSTypeGardenLinux is a constant for the FIPS-enabled Garden Linux extension OS type.
	OSTypeGardenLinuxFips = "gardenlinux-fips"
)
