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

	// OSTypeGardenLinux is a constant for the Garden Linux extension OS type.
	OSTypeGardenLinux = "gardenlinux"
)
