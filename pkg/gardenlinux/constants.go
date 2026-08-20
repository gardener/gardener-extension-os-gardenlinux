// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gardenlinux

import (
	"embed"
)

var (
	//go:embed scripts/*
	Templates embed.FS

	ScriptPermissions = uint32(0755)
)

const (
	// ScriptLocation is the location that Gardener configuration scripts end up on Garden Linux
	ScriptLocation = "/opt/gardener/bin"

	// PathEtcSetupHook is the path of the etc-setup hook script on the worker node.
	// GardenLinux executes all executables in this directory via run-parts after the /etc overlay is
	// wiped during a major OS version upgrade (wipe-based in-place update). The filename has no
	// extension so that Debian run-parts (default regex ^[a-zA-Z0-9_-]+$) does not skip it.
	PathEtcSetupHook = "/var/lib/gardenlinux/etc-setup-hooks/00-gardener"

	// OSTypeGardenLinux is a constant for the Garden Linux extension OS type.
	OSTypeGardenLinux = "gardenlinux"

	// OSTypeGardenLinux is a constant for the FIPS-enabled Garden Linux extension OS type.
	OSTypeGardenLinuxFips = "gardenlinux-fips"
)
