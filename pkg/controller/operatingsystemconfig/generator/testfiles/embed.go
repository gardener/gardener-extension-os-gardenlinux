// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package testfiles

import (
	"embed"
)

// Files contains the contents of the testfiles directory
//
//go:embed cloud-* containerd-* docker-* memoryone-*
var Files embed.FS
