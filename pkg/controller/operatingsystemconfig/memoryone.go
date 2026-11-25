// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig

import (
	"fmt"
	"strings"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	memoryonegardenlinux "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/memoryonegardenlinux/v1alpha1"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/memoryone"
)

const (
	memoryTopology = "mem_topology"
	systemMemory   = "system_memory"
)

func wrapIntoMemoryOneHeaderAndFooter(osc *extensionsv1alpha1.OperatingSystemConfig, in string) (string, error) {
	config, err := memoryone.Configuration(osc)
	if err != nil {
		return "", err
	}

	memoryOneConfiguration := vsmpConfigString(config)

	out := `Content-Type: multipart/mixed; boundary="==BOUNDARY=="
MIME-Version: 1.0
--==BOUNDARY==
Content-Type: text/x-vsmp; section=vsmp

` + memoryOneConfiguration + `
--==BOUNDARY==
Content-Type: text/x-shellscript

` + in + `
--==BOUNDARY==--
`

	return out, nil
}

func vsmpConfigString(config *memoryonegardenlinux.OperatingSystemConfiguration) string {
	var vsmpConfiguration map[string]string
	var configStringBuilder strings.Builder

	if config != nil && config.VsmpConfiguration != nil {
		vsmpConfiguration = config.VsmpConfiguration
		// TODO: put stripSemicolaAndNewlines down into the StringBuilder-Fprintf and remove this loop once we end support for legacy values
		// this is required as we do not want to allow injecting key-value pairs with semicola in the new parameter map
		// but need to retain the previous behaviour for the legacy configuration style
		for k, v := range vsmpConfiguration {
			vsmpConfiguration[k] = stripSemicolaAndNewlines(v)
		}
	} else {
		vsmpConfiguration = make(map[string]string, 2)
	}

	if _, ok := vsmpConfiguration[memoryTopology]; !ok {
		vsmpConfiguration[memoryTopology] = "2"
	}

	if _, ok := vsmpConfiguration[systemMemory]; !ok {
		vsmpConfiguration[systemMemory] = "6x"
	}

	// TODO: remove these once the transition to VsmpConfiguration map[string]string is complete
	if config != nil {
		if config.SystemMemory != nil {
			vsmpConfiguration[systemMemory] = *config.SystemMemory
		}

		if config.MemoryTopology != nil {
			vsmpConfiguration[memoryTopology] = *config.MemoryTopology
		}
	}
	// end TODO

	for k, v := range vsmpConfiguration {
		// values can be arbitrary strings and must not be sanitized
		fmt.Fprintf(&configStringBuilder, "%s=%s\n", stripSemicolaAndNewlines(k), v)
	}

	return configStringBuilder.String()
}

func stripNewLines(s string) string {
	before, _, found := strings.Cut(s, "\n")
	if found {
		return before
	}

	return s
}

func stripSemicola(s string) string {
	before, _, found := strings.Cut(s, ";")
	if found {
		return before
	}
	return s
}

func stripSemicolaAndNewlines(s string) string {
	return stripSemicola(stripNewLines(s))
}
