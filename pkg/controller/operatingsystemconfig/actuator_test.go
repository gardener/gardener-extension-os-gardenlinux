// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig_test

import (
	"context"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/gardener/gardener-extension-os-gardenlinux/pkg/controller/operatingsystemconfig"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/memoryone"
)

var _ = Describe("Actuator", func() {
	var (
		ctx        = context.TODO()
		log        = logr.Discard()
		fakeClient client.Client
		mgr        manager.Manager

		osc      *extensionsv1alpha1.OperatingSystemConfig
		actuator operatingsystemconfig.Actuator
	)

	BeforeEach(func() {
		fakeClient = fakeclient.NewClientBuilder().Build()
		mgr = test.FakeManager{Client: fakeClient}
		actuator = NewActuator(mgr)

		osc = &extensionsv1alpha1.OperatingSystemConfig{
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type: gardenlinux.OSTypeGardenLinux,
				},
				Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
				Units:   []extensionsv1alpha1.Unit{{Name: "some-unit", Content: ptr.To("foo")}},
				Files:   []extensionsv1alpha1.File{{Path: "/some/file", Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: "bar"}}}},
			},
		}
	})

	When("purpose is 'provision'", func() {
		expectedUserData := `#!/bin/bash
if [ -f "/var/lib/osc/provision-osc-applied" ]; then
  echo "Provision OSC already applied, exiting..."
  exit 0
fi

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

mkdir -p "/some"

cat << EOF | base64 -d > "/some/file"
YmFy
EOF


cat << EOF | base64 -d > "/etc/systemd/system/some-unit"
Zm9v
EOF
grep -sq "^nfsd$" /etc/modules || echo "nfsd" >>/etc/modules
modprobe nfsd
nslookup $(hostname) || systemctl restart systemd-networkd

systemctl daemon-reload
systemctl enable containerd && systemctl restart containerd
systemctl enable 'some-unit' && systemctl restart --no-block 'some-unit'


mkdir -p /var/lib/osc
touch /var/lib/osc/provision-osc-applied
`

		When("OS type is 'gardenlinux'", func() {
			Describe("#Reconcile", func() {
				It("should not return an error", func() {
					userData, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(userData)).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
				})
			})
		})

		When("OS type is 'memoryone-gardenlinux'", func() {
			BeforeEach(func() {
				osc.Spec.Type = memoryone.OSTypeMemoryOneGardenLinux
				osc.Spec.ProviderConfig = &runtime.RawExtension{Raw: []byte(`apiVersion: memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1
kind: OperatingSystemConfiguration
memoryTopology: "2"
systemMemory: "6x"`)}
			})

			Describe("#Reconcile", func() {
				It("should not return an error", func() {
					userData, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(userData)).To(Equal(`Content-Type: multipart/mixed; boundary="==BOUNDARY=="
MIME-Version: 1.0
--==BOUNDARY==
Content-Type: text/x-vsmp; section=vsmp
system_memory=6x
mem_topology=2
--==BOUNDARY==
Content-Type: text/x-shellscript
` + expectedUserData + `
--==BOUNDARY==`))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
				})
			})
		})
	})

	When("purpose is 'reconcile'", func() {
		BeforeEach(func() {
			osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
		})

		Describe("#Reconcile", func() {
			It("should not return usersdata for purpose reconcile", func() {
				userData, _, _, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())
				Expect(userData).To(BeEmpty())
			})

			It("should add one empty additional unit for containerd", func() {
				_, units, files, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(units)).To(Equal(1))
				Expect(units).To(ContainElement(
					extensionsv1alpha1.Unit{
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
				))
				Expect(files).To(BeEmpty())
			})
		})
	})
})
