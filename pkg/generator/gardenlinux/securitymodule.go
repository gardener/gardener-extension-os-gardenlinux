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
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// defaultLsm is the Linux security module to fall back to
var defaultLsm = v1alpha1.LinuxSecurityModule(v1alpha1.LsmAppArmor)

var lsmUnitContent string = `[Unit]
Description=Configure lsm for Gardener
After=cloud-config-downloader.service
Before=gardener-restart-system.service kubelet.service

[Install]
WantedBy=multi-user.target

[Service]
Type=oneshot
ExecStart=/opt/gardener/bin/configure_lsm.sh
RemainAfterExit=true
StandardOutput=journal
`

func ConfigureLinuxSecurityModule(osc *extensionsv1alpha1.OperatingSystemConfig, decoder runtime.Decoder) (*generator.File, *generator.Unit, error) {
	providerConfig := osc.Spec.ProviderConfig
	lsm := defaultLsm

	if providerConfig != nil {
		obj := &v1alpha1.OperatingSystemConfiguration{}

		if _, _, err := decoder.Decode(providerConfig.Raw, nil, obj); err != nil {
			return nil, nil, fmt.Errorf("failed to decode provider config: %+v", err)
		}

		if len(obj.LinuxSecurityModule) != 0 {
			lsm = obj.LinuxSecurityModule
		}
	}

	config := map[string]interface{}{
		"linuxSecurityModule": lsm,
	}

	var buff bytes.Buffer
	lsmScriptTemplate, err := templates.ReadFile(filepath.Join("templates", "configure_lsm.sh.tpl"))
	if err != nil {
		return nil, nil, err
	}
	t, err := template.New("lsmScript").Parse(string(lsmScriptTemplate))
	if err != nil {
		return nil, nil, err
	}
	if err = t.Execute(&buff, config); err != nil {
		return nil, nil, err
	}

	lsmScript := &generator.File{
		Path:        filepath.Join(scriptLocation, "configure_lsm.sh"),
		Content:     buff.Bytes(),
		Permissions: &scriptPermissions,
	}

	lsmUnit := &generator.Unit{
		Name:    "gardener-configure-lsm.service",
		Content: []byte(lsmUnitContent),
	}

	return lsmScript, lsmUnit, nil
}
