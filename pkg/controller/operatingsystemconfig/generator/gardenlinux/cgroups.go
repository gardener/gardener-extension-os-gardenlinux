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
	containerdConfigureScriptContent, err := gardenlinux.Templates.ReadFile(filepath.Join("scripts", "containerd_cgroup_driver.sh"))
	if err != nil {
		return nil, nil, err
	}
	containerdConfigureScript := &generator.File{
		Path:        filepath.Join(gardenlinux.ScriptLocation, "containerd_cgroup_driver.sh"),
		Content:     containerdConfigureScriptContent,
		Permissions: &gardenlinux.ScriptPermissions,
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
	kubeletConfigureScriptContent, err := gardenlinux.Templates.ReadFile(filepath.Join("scripts", "kubelet_cgroup_driver.sh"))
	if err != nil {
		return nil, nil, err
	}
	kubeletConfigureScript := &generator.File{
		Path:        filepath.Join(gardenlinux.ScriptLocation, "kubelet_cgroup_driver.sh"),
		Content:     kubeletConfigureScriptContent,
		Permissions: &gardenlinux.ScriptPermissions,
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
