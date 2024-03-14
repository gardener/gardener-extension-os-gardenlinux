// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gardenlinux

import (
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
)

var (
	fileList = []string{
		"g_functions.sh",
	}
)

// GetAdditionalScripts returns additional scripts that were provided as raw files in the embedded fs
func GetAdditionalScripts() ([]*generator.File, error) {
	files := []*generator.File{}

	for _, f := range fileList {
		scriptContent, err := gardenlinux.Templates.ReadFile(filepath.Join("scripts", f))
		if err != nil {
			return nil, err
		}

		additionalScript := &generator.File{
			Path:        filepath.Join(gardenlinux.ScriptLocation, f),
			Content:     scriptContent,
			Permissions: &gardenlinux.ScriptPermissions,
		}

		files = append(files, additionalScript)
	}

	return files, nil
}
