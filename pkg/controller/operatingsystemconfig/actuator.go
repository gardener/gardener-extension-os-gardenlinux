// Copyright 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package operatingsystemconfig

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	oscommonactuator "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/controller/operatingsystemconfig/generator"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/memoryone"
)

type actuator struct {
	client               client.Client
	useGardenerNodeAgent bool
}

// NewActuator creates a new Actuator that updates the status of the handled OperatingSystemConfig resources.
func NewActuator(mgr manager.Manager, useGardenerNodeAgent bool) operatingsystemconfig.Actuator {
	return &actuator{
		client:               mgr.GetClient(),
		useGardenerNodeAgent: useGardenerNodeAgent,
	}
}

func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, []string, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	cloudConfig, command, err := oscommonactuator.CloudConfigFromOperatingSystemConfig(ctx, log, a.client, osc, generator.CloudInitGenerator())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("could not generate cloud config: %w", err)
	}

	switch purpose := osc.Spec.Purpose; purpose {
	case extensionsv1alpha1.OperatingSystemConfigPurposeProvision:
		if !a.useGardenerNodeAgent {
			return cloudConfig, command, oscommonactuator.OperatingSystemConfigUnitNames(osc), oscommonactuator.OperatingSystemConfigFilePaths(osc), nil, nil, nil
		}
		userData, err := a.handleProvisionOSC(ctx, osc)
		return []byte(userData), nil, nil, nil, nil, nil, err

	case extensionsv1alpha1.OperatingSystemConfigPurposeReconcile:
		extensionUnits, extensionFiles, err := a.handleReconcileOSC(osc)
		return cloudConfig, command, oscommonactuator.OperatingSystemConfigUnitNames(osc), oscommonactuator.OperatingSystemConfigFilePaths(osc), extensionUnits, extensionFiles, err

	default:
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("unknown purpose: %s", purpose)
	}
}

func (a *actuator) Delete(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.OperatingSystemConfig) error {
	return nil
}

func (a *actuator) Migrate(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) error {
	return a.Delete(ctx, log, osc)
}

func (a *actuator) ForceDelete(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) error {
	return a.Delete(ctx, log, osc)
}

func (a *actuator) Restore(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, []string, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	return a.Reconcile(ctx, log, osc)
}

func (a *actuator) handleProvisionOSC(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (string, error) {
	writeFilesToDiskScript, err := operatingsystemconfig.FilesToDiskScript(ctx, a.client, osc.Namespace, osc.Spec.Files)
	if err != nil {
		return "", err
	}
	writeUnitsToDiskScript := operatingsystemconfig.UnitsToDiskScript(osc.Spec.Units)

	script := `#!/bin/bash
if [ ! -s /etc/containerd/config.toml ]; then
  mkdir -p /etc/containerd/
  containerd config default > /etc/containerd/config.toml
  chmod 0644 /etc/containerd/config.toml
fi

mkdir -p /etc/systemd/system/containerd.service.d
cat <<EOF > /etc/systemd/system/containerd.service.d/11-exec_config.conf
[Service]
ExecStart=
ExecStart=/usr/bin/containerd --config=/etc/containerd/config.toml
EOF
chmod 0644 /etc/systemd/system/containerd.service.d/11-exec_config.conf
` + writeFilesToDiskScript + `
` + writeUnitsToDiskScript + `
grep -sq "^nfsd$" /etc/modules || echo "nfsd" >>/etc/modules
modprobe nfsd
nslookup $(hostname) || systemctl restart systemd-networkd

systemctl daemon-reload
systemctl enable containerd && systemctl restart containerd
systemctl enable docker && systemctl restart docker
systemctl enable gardener-node-init && systemctl restart gardener-node-init`

	if osc.Spec.Type == memoryone.OSTypeMemoryOneGardenLinux {
		return wrapIntoMemoryOneHeaderAndFooter(osc, script)
	}

	return script, nil
}

func wrapIntoMemoryOneHeaderAndFooter(osc *extensionsv1alpha1.OperatingSystemConfig, in string) (string, error) {
	config, err := memoryone.Configuration(osc)
	if err != nil {
		return "", err
	}

	out := `Content-Type: multipart/mixed; boundary="==BOUNDARY=="
MIME-Version: 1.0
--==BOUNDARY==
Content-Type: text/x-vsmp; section=vsmp`

	if config != nil && config.SystemMemory != nil {
		out += fmt.Sprintf(`
system_memory=%s`, *config.SystemMemory)
	}
	if config != nil && config.MemoryTopology != nil {
		out += fmt.Sprintf(`
mem_topology=%s`, *config.MemoryTopology)
	}

	out += `
--==BOUNDARY==
Content-Type: text/x-shellscript
` + in + `
--==BOUNDARY==`

	return out, nil
}

var (
	scriptContentGFunctions             []byte
	scriptContentKubeletCGroupDriver    []byte
	scriptContentContainerdCGroupDriver []byte
)

func init() {
	var err error

	scriptContentGFunctions, err = gardenlinux.Templates.ReadFile(filepath.Join("scripts", "g_functions.sh"))
	utilruntime.Must(err)
	scriptContentKubeletCGroupDriver, err = gardenlinux.Templates.ReadFile(filepath.Join("scripts", "kubelet_cgroup_driver.sh"))
	utilruntime.Must(err)
	scriptContentContainerdCGroupDriver, err = gardenlinux.Templates.ReadFile(filepath.Join("scripts", "containerd_cgroup_driver.sh"))
	utilruntime.Must(err)
}

func (a *actuator) handleReconcileOSC(_ *extensionsv1alpha1.OperatingSystemConfig) ([]extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	var (
		extensionUnits []extensionsv1alpha1.Unit
		extensionFiles []extensionsv1alpha1.File
	)

	filePathFunctionsHelperScript := filepath.Join(gardenlinux.ScriptLocation, "g_functions.sh")
	extensionFiles = append(extensionFiles, extensionsv1alpha1.File{
		Path:        filePathFunctionsHelperScript,
		Content:     extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: string(scriptContentGFunctions)}},
		Permissions: &gardenlinux.ScriptPermissions,
	})

	// add scripts and dropins for kubelet
	filePathKubeletCGroupDriverScript := filepath.Join(gardenlinux.ScriptLocation, "kubelet_cgroup_driver.sh")
	extensionFiles = append(extensionFiles, extensionsv1alpha1.File{
		Path:        filePathKubeletCGroupDriverScript,
		Content:     extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: string(scriptContentKubeletCGroupDriver)}},
		Permissions: &gardenlinux.ScriptPermissions,
	})
	extensionUnits = append(extensionUnits, extensionsv1alpha1.Unit{
		Name: "kubelet.service",
		DropIns: []extensionsv1alpha1.DropIn{{
			Name: "10-configure-cgroup-driver.conf",
			Content: `[Service]
ExecStartPre=` + filePathKubeletCGroupDriverScript + `
`,
		}},
		FilePaths: []string{filePathFunctionsHelperScript, filePathKubeletCGroupDriverScript},
	})

	// add scripts and dropins for containerd if activated
	filePathContainerdCGroupDriverScript := filepath.Join(gardenlinux.ScriptLocation, "containerd_cgroup_driver.sh")
	extensionFiles = append(extensionFiles, extensionsv1alpha1.File{
		Path:        filePathContainerdCGroupDriverScript,
		Content:     extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: string(scriptContentContainerdCGroupDriver)}},
		Permissions: &gardenlinux.ScriptPermissions,
	})
	extensionUnits = append(extensionUnits, extensionsv1alpha1.Unit{
		Name: "containerd.service",
		DropIns: []extensionsv1alpha1.DropIn{{
			Name: "10-configure-cgroup-driver.conf",
			Content: `[Service]
ExecStartPre=` + filePathContainerdCGroupDriverScript + `
`,
		}},
		FilePaths: []string{filePathFunctionsHelperScript, filePathContainerdCGroupDriverScript},
	})

	return extensionUnits, extensionFiles, nil
}
