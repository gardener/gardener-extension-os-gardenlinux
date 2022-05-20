// Copyright (c) 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gardenlinux

import (
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
)

var restartUnitContent string = `[Unit]
Description=Optionally restart system to apply Gardener specific OS configuration
Documentation=https://github.com/gardenlinux/gardenlinux/docs/gardener-kernel-restart.md
After=gardener-configure-cgroups.service garderner-configure-lsm.service
Before=kubelet.service

[Install]
WantedBy=multi-user.target

[Service]
Type=oneshot
ExecStart=/opt/gardener/bin/restart_system.sh
RemainAfterExit=true
StandardOutput=journal
`

func PlaceRestartUnit() (*generator.File, *generator.Unit, error) {
	restartScriptContent, err := templates.ReadFile(filepath.Join("files", "restart_system.sh"))
	if err != nil {
		return nil, nil, err
	}

	restartScript := &generator.File{
		Path:        filepath.Join(scriptLocation, "restart_system.sh"),
		Content:     restartScriptContent,
		Permissions: &scriptPermissions,
	}

	restartUnit := &generator.Unit{
		Name:    "gardener-restart-system.service",
		Content: []byte(restartUnitContent),
	}

	return restartScript, restartUnit, nil
}
