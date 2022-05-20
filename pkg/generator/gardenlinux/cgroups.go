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
	"context"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux/v1alpha1"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/version"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// defaultCgroup is the cgroup version to fall back to
	defaultCgroup = v1alpha1.CgroupVersion(v1alpha1.CgroupVersionV1)

	// cgroupUnitContent is the content of the systemd unit to start the cgroup configuration script
	cgroupUnitContent string = `[Unit]
Description=Configure cgroup version for Gardener
After=cloud-config-downloader.service
Before=gardener-restart-system.service kubelet.service

[Install]
WantedBy=multi-user.target

[Service]
Type=oneshot
ExecStart=/opt/gardener/bin/configure_cgroups.sh
RemainAfterExit=true
StandardOutput=journal
`

	// containerdDropInContent is the content of the systemd drop-in to configure the cgroup driver for containerd
	containerdDropInContent string = `[Service]
ExecStartPre=/opt/gardener/bin/containerd_cgroup_driver.sh
`

	// kubeletDropInContent is the content of the systemd drop-in to configure the cgroup driver for kubelet
	kubeletDropInContent string = `[Service]
ExecStartPre=/opt/gardener/bin/kubelet_cgroup_driver.sh
`
)

func ConfigureCgroups(ctx context.Context, client client.Client, osc *extensionsv1alpha1.OperatingSystemConfig, decoder runtime.Decoder) (*generator.File, *generator.Unit, error) {
	cgroupVersion, err := extractCgroupVersion(ctx, client, osc, decoder)
	if err != nil {
		return nil, nil, err
	}

	config := map[string]interface{}{
		"cgroupVersion": cgroupVersion,
	}

	var buff bytes.Buffer
	cgroupScriptTemplate, err := templates.ReadFile(filepath.Join("templates", "configure_cgroups.sh.tpl"))
	if err != nil {
		return nil, nil, err
	}
	t, err := template.New("cgroupScript").Parse(string(cgroupScriptTemplate))
	if err != nil {
		return nil, nil, err
	}
	if err = t.Execute(&buff, config); err != nil {
		return nil, nil, err
	}

	cgroupScript := &generator.File{
		Path:        filepath.Join(scriptLocation, "configure_cgroups.sh"),
		Content:     buff.Bytes(),
		Permissions: &scriptPermissions,
	}

	cgroupUnit := &generator.Unit{
		Name:    "gardener-configure-cgroups.service",
		Content: []byte(cgroupUnitContent),
	}

	return cgroupScript, cgroupUnit, nil
}

func extractCgroupVersion(ctx context.Context, client client.Client, osc *extensionsv1alpha1.OperatingSystemConfig, decoder runtime.Decoder) (*v1alpha1.CgroupVersion, error) {
	providerConfig := osc.Spec.ProviderConfig

	if providerConfig != nil {
		obj := &v1alpha1.OperatingSystemConfiguration{}

		if _, _, err := decoder.Decode(providerConfig.Raw, nil, obj); err != nil {
			return nil, fmt.Errorf("failed to decode provider config: %+v", err)
		}

		shoot, err := extensionscontroller.GetShoot(ctx, client, osc.Namespace)
		if err != nil {
			return nil, err
		}

		shootKubernetesAtLeast119, err := version.CompareVersions(shoot.Spec.Kubernetes.Version, ">=", "1.19")
		if err != nil {
			return nil, err
		}

		if shootKubernetesAtLeast119 {
			return &obj.CgroupVersion, nil
		}
	}

	return &defaultCgroup, nil
}

func ContainerRuntimesCgroupDrivers() ([]*generator.File, []*generator.Unit, error) {
	containerdConfigureScriptContent, err := templates.ReadFile(filepath.Join("files", "containerd_cgroup_driver.sh"))
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

	kubeletConfigureScriptContent, err := templates.ReadFile(filepath.Join("files", "kubelet_cgroup_driver.sh"))
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
	return []*generator.File{containerdConfigureScript, kubeletConfigureScript}, []*generator.Unit{containerdDropin, kubeletDropin}, nil
}
