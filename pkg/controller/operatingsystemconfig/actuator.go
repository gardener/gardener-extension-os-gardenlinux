// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/go-logr/logr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/memoryone"
)

type actuator struct {
	client client.Client
}

// NewActuator creates a new Actuator that updates the status of the handled OperatingSystemConfig resources.
func NewActuator(mgr manager.Manager) operatingsystemconfig.Actuator {
	return &actuator{
		client: mgr.GetClient(),
	}
}

func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, *extensionsv1alpha1.InPlaceUpdatesStatus, error) {
	switch purpose := osc.Spec.Purpose; purpose {
	case extensionsv1alpha1.OperatingSystemConfigPurposeProvision:
		userData, err := a.handleProvisionOSC(ctx, osc)
		return []byte(userData), nil, nil, nil, err

	case extensionsv1alpha1.OperatingSystemConfigPurposeReconcile:
		extensionUnits, extensionFiles, inPlaceUpdates, err := a.handleReconcileOSC(osc)
		return nil, extensionUnits, extensionFiles, inPlaceUpdates, err

	default:
		return nil, nil, nil, nil, fmt.Errorf("unknown purpose: %s", purpose)
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

func (a *actuator) Restore(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, *extensionsv1alpha1.InPlaceUpdatesStatus, error) {
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
`
	for _, unit := range osc.Spec.Units {
		script += fmt.Sprintf(`systemctl enable '%s' && systemctl restart --no-block '%s'
`, unit.Name, unit.Name)
	}

	// The provisioning script must run only once.
	script = operatingsystemconfig.WrapProvisionOSCIntoOneshotScript(script)

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

var scriptContentInPlaceUpdate []byte

func init() {
	var err error

	scriptContentInPlaceUpdate, err = gardenlinux.Templates.ReadFile(filepath.Join("scripts", "inplace-update.sh"))
	utilruntime.Must(err)
}

func (a *actuator) handleReconcileOSC(osc *extensionsv1alpha1.OperatingSystemConfig) ([]extensionsv1alpha1.Unit, []extensionsv1alpha1.File, *extensionsv1alpha1.InPlaceUpdatesStatus, error) {
	var (
		extensionUnits []extensionsv1alpha1.Unit
		extensionFiles []extensionsv1alpha1.File
		inPlaceUpdates *extensionsv1alpha1.InPlaceUpdatesStatus
	)

	extensionUnits = []extensionsv1alpha1.Unit{
		{
			Name: "containerd.service",
			DropIns: []extensionsv1alpha1.DropIn{
				{
					Name: "override.conf",
					Content: `[Service]
LimitMEMLOCK=67108864
LimitNOFILE=1048576`,
				},
			},
		},
	}

	if osc.Spec.InPlaceUpdates != nil {
		filePathOSUpdateScript := filepath.Join(gardenlinux.ScriptLocation, "inplace-update.sh")
		extensionFiles = append(extensionFiles, extensionsv1alpha1.File{
			Path: filePathOSUpdateScript,
			Content: extensionsv1alpha1.FileContent{
				Inline: &extensionsv1alpha1.FileContentInline{
					Data:     utils.EncodeBase64(scriptContentInPlaceUpdate),
					Encoding: "b64",
				},
			},
			Permissions: &gardenlinux.ScriptPermissions,
		})

		inPlaceUpdates = &extensionsv1alpha1.InPlaceUpdatesStatus{
			OSUpdate: &extensionsv1alpha1.OSUpdate{
				Command: filePathOSUpdateScript,
				Args:    []string{osc.Spec.InPlaceUpdates.OperatingSystemVersion},
			},
		}
	}

	return extensionUnits, extensionFiles, inPlaceUpdates, nil
}
