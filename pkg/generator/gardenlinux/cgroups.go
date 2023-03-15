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

var (
	// containerdDropInContent is the content of the systemd drop-in to configure the cgroup driver for containerd
	containerdDropInContent string = `[Service]
ExecStartPre=/opt/gardener/bin/containerd_cgroup_driver.sh
`

	// kubeletDropInContent is the content of the systemd drop-in to configure the cgroup driver for kubelet
	kubeletDropInContent string = `[Service]
ExecStartPre=/opt/gardener/bin/kubelet_cgroup_driver.sh
`
)

func ContainerdCgroupDriver() ([]*generator.File, []*generator.Unit, error) {
	containerdConfigureScriptContent, err := templates.ReadFile(filepath.Join("scripts", "containerd_cgroup_driver.sh"))
	if err != nil {
		return nil, nil, err
	}
	containerdConfigureScript := &generator.File{
		Path:        filepath.Join(scriptLocation, "containerd_cgroup_driver.sh"),
		Content:     containerdConfigureScriptContent,
		Permissions: &scriptPermissions,
	}
	containerdDropin := &generator.Unit{
		Name: "containerd.service",
		DropIns: []*generator.DropIn{
			{
				Name:    "10-configure-cgroup-driver.conf",
				Content: []byte(containerdDropInContent),
			},
		},
	}

	return []*generator.File{containerdConfigureScript}, []*generator.Unit{containerdDropin}, nil
}

func KubeletCgroupDriver() ([]*generator.File, []*generator.Unit, error) {
	kubeletConfigureScriptContent, err := templates.ReadFile(filepath.Join("scripts", "kubelet_cgroup_driver.sh"))
	if err != nil {
		return nil, nil, err
	}
	kubeletConfigureScript := &generator.File{
		Path:        filepath.Join(scriptLocation, "kubelet_cgroup_driver.sh"),
		Content:     kubeletConfigureScriptContent,
		Permissions: &scriptPermissions,
	}
	kubeletDropin := &generator.Unit{
		Name: "kubelet.service",
		DropIns: []*generator.DropIn{
			{
				Name:    "10-configure-cgroup-driver.conf",
				Content: []byte(kubeletDropInContent),
			},
		},
	}

	return []*generator.File{kubeletConfigureScript}, []*generator.Unit{kubeletDropin}, nil
}
