// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig_test

import (
	"context"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	nodeagentconfigv1alpha1 "github.com/gardener/gardener/pkg/apis/config/nodeagent/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	nodeagentcomponent "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/nodeagent"
	"github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/rootcertificates"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	memoryonegardenlinuxv1alpha1 "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/memoryonegardenlinux/v1alpha1"
	. "github.com/gardener/gardener-extension-os-gardenlinux/pkg/controller/operatingsystemconfig"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/memoryone"
)

var codec runtime.Codec

func init() {
	scheme := runtime.NewScheme()
	runtimeutils.Must(memoryonegardenlinuxv1alpha1.AddToScheme(scheme))
	codec = serializer.NewCodecFactory(scheme, serializer.EnableStrict).LegacyCodec(memoryonegardenlinuxv1alpha1.SchemeGroupVersion)
}

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

systemctl daemon-reload
systemctl enable containerd && systemctl restart containerd
systemctl enable 'some-unit' && systemctl restart --no-block 'some-unit'


mkdir -p /var/lib/osc
touch /var/lib/osc/provision-osc-applied
`
		DescribeTableSubtree("OSC type is ", func(osctype string) {
			It("should not return an error", func() {
				osc.Spec.Type = osctype
				userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(userData)).To(Equal(expectedUserData))
				Expect(extensionUnits).To(BeEmpty())
				Expect(extensionFiles).To(BeEmpty())
				Expect(inplaceUpdateStatus).To(BeNil())
			})
		},
			Entry("gardenlinux", "gardenlinux"),
			Entry("gardenlinux-fips", "gardenlinux-fips"),
		)

		When("OS type is 'memoryone-gardenlinux'", func() {
			var memoryOneConfiguration memoryonegardenlinuxv1alpha1.OperatingSystemConfiguration

			BeforeEach(func() {
				memoryOneConfiguration = memoryonegardenlinuxv1alpha1.OperatingSystemConfiguration{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1",
						Kind:       "OperatingSystemConfiguration",
					},
				}

				osc.Spec.Type = memoryone.OSTypeMemoryOneGardenLinux
			})

			When("Legacy fields are used", func() {
				It("should use default values for the system_memory and mem_topology", func() {
					osc.Spec.ProviderConfig = nil

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "2",
						"system_memory": "6x",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
				})

				It("should use custom values for system_memory and mem_topology", func() {
					memoryOneConfiguration.MemoryTopology = ptr.To("4")
					memoryOneConfiguration.SystemMemory = ptr.To("8x")
					Expect(encodeMemoryOneConfigurationIntoOsc(codec, osc, &memoryOneConfiguration)).To(Succeed())

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "4",
						"system_memory": "8x",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
				})

				It("should allow injecting additional key-value pairs by semicola", func() {
					memoryOneConfiguration.MemoryTopology = ptr.To("4; foo=bar")
					memoryOneConfiguration.SystemMemory = ptr.To("8x")
					Expect(encodeMemoryOneConfigurationIntoOsc(codec, osc, &memoryOneConfiguration)).To(Succeed())

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "4; foo=bar",
						"system_memory": "8x",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
				})
			})

			When("MemoryOne configuration map is used", func() {
				It("Should include arbitrary configuration values in vSMP config", func() {
					memoryOneConfiguration.VsmpConfiguration = map[string]string{
						"foo": "bar",
						"abc": "xyz",
					}
					Expect(encodeMemoryOneConfigurationIntoOsc(codec, osc, &memoryOneConfiguration)).To(Succeed())

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "2",
						"system_memory": "6x",
						"foo":           "bar",
						"abc":           "xyz",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
				})

				It("Should not allow injecting additional key-value pairs by semicola", func() {
					memoryOneConfiguration.VsmpConfiguration = map[string]string{
						"foo":        "bar; foobar: barfoo",
						"newlinefoo": "bar\nwith\nnew\nline",
						"abc":        "xyz",
					}
					Expect(encodeMemoryOneConfigurationIntoOsc(codec, osc, &memoryOneConfiguration)).To(Succeed())

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "2",
						"system_memory": "6x",
						"foo":           "bar",
						"newlinefoo":    "bar",
						"abc":           "xyz",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
				})

				It("Should allow quoted values", func() {
					memoryOneConfiguration.VsmpConfiguration = map[string]string{
						"quoted": "\"12:34:56:78:90:ab:cd:ef\"",
					}
					Expect(encodeMemoryOneConfigurationIntoOsc(codec, osc, &memoryOneConfiguration)).To(Succeed())

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "2",
						"system_memory": "6x",
						"quoted":        "\"12:34:56:78:90:ab:cd:ef\"",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
				})

				It("Should give priority to legacy values", func() {
					memoryOneConfiguration.MemoryTopology = ptr.To("3")
					memoryOneConfiguration.SystemMemory = ptr.To("7x")
					memoryOneConfiguration.VsmpConfiguration = map[string]string{
						"mem_topology":  "5",
						"system_memory": "13x",
					}
					Expect(encodeMemoryOneConfigurationIntoOsc(codec, osc, &memoryOneConfiguration)).To(Succeed())

					userData, extensionUnits, extensionFiles, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					vSmpConfig, decodedUserData := decodeVsmpUserData(string(userData))

					Expect(vSmpConfig).To(BeEquivalentTo(map[string]string{
						"mem_topology":  "3",
						"system_memory": "7x",
					}))
					Expect(decodedUserData).To(Equal(expectedUserData))
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
					Expect(inplaceUpdateStatus).To(BeNil())
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
				userData, _, _, _, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())
				Expect(userData).To(BeEmpty())
			})

			Context("In-Place Updates", func() {
				BeforeEach(func() {
					osc.Spec.InPlaceUpdates = &extensionsv1alpha1.InPlaceUpdates{
						OperatingSystemVersion: "1.0.0-inplace",
					}
					osc.Spec.Units = []extensionsv1alpha1.Unit{
						{Name: "gardener-node-agent.service", Content: new("[Unit]\nDescription=GNA")},
						{Name: "kubelet.service", Content: new("[Unit]\nDescription=kubelet")},
						{Name: "no-content.service"}, // units without content
					}
				})

				It("should return InPlaceUpdatesStatus with the OS update command", func() {
					_, _, _, inplaceUpdateStatus, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					Expect(inplaceUpdateStatus).To(Equal(&extensionsv1alpha1.InPlaceUpdatesStatus{
						OSUpdate: &extensionsv1alpha1.OSUpdate{
							Command: "/opt/gardener/bin/inplace-update.sh",
							Args:    []string{"1.0.0"},
						},
					}))
				})

				It("should deliver the inplace-update.sh script and the etc-setup hook file", func() {
					_, _, files, _, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					paths := make([]string, 0, len(files))
					for _, f := range files {
						paths = append(paths, f.Path)
					}
					Expect(paths).To(ContainElements(
						"/opt/gardener/bin/inplace-update.sh",
						gardenlinux.PathEtcSetupHook,
					))
				})

				It("should restore only the init-equivalent /etc state and hand off to gardener-node-agent", func() {
					_, _, files, _, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					var hookFile *extensionsv1alpha1.File
					for i := range files {
						if files[i].Path == gardenlinux.PathEtcSetupHook {
							hookFile = &files[i]
							break
						}
					}
					Expect(hookFile).NotTo(BeNil())
					Expect(hookFile.Content.Inline).NotTo(BeNil())
					Expect(hookFile.Content.Inline.Encoding).To(Equal("b64"))

					decoded, err := base64.StdEncoding.DecodeString(hookFile.Content.Inline.Data)
					Expect(err).NotTo(HaveOccurred())

					script := string(decoded)
					Expect(script).To(ContainSubstring("#!/usr/bin/env bash"))

					// The gardener-node-agent unit is restored (base64-encoded content to avoid heredoc injection) and started.
					Expect(script).To(ContainSubstring(`> "/etc/systemd/system/` + nodeagentconfigv1alpha1.UnitName + `"`))
					Expect(script).To(ContainSubstring(base64.StdEncoding.EncodeToString([]byte("[Unit]\nDescription=GNA"))))
					Expect(script).To(ContainSubstring("systemctl enable " + nodeagentconfigv1alpha1.UnitName))

					// containerd is brought up so gardener-node-agent can start.
					Expect(script).To(ContainSubstring("containerd config default"))
					Expect(script).To(ContainSubstring("11-exec_config.conf"))
					Expect(script).To(ContainSubstring("systemctl daemon-reload"))
					Expect(script).To(ContainSubstring("systemctl enable --now containerd.service"))

					// The OS system trust store is rebuilt from the surviving /var CA sources so gardener-node-agent
					// can pull its own image from a CA-fronted/private registry before it re-applies the OSC.
					Expect(script).To(ContainSubstring(rootcertificates.PathUpdateLocalCACertificates))

					// Prerequisites that make gardener-node-agent startup impossible are verified and fail loudly.
					Expect(script).To(ContainSubstring(`NODE_AGENT_BINARY="` + nodeagentcomponent.PathBinary + `"`))
					Expect(script).To(ContainSubstring(`NODE_AGENT_CONFIG_DIR="` + nodeagentconfigv1alpha1.BaseDir + `"`))
					Expect(script).To(ContainSubstring(`[ ! -x "${NODE_AGENT_BINARY}" ]`))
					Expect(script).To(ContainSubstring(`ls "${NODE_AGENT_CONFIG_DIR}"/config-*.yaml`))
					// gardener-node-agent is started with a blocking start so activation failures surface.
					Expect(script).To(ContainSubstring("systemctl restart " + nodeagentconfigv1alpha1.UnitName))

					// The hook does NOT restore the rest of the OperatingSystemConfig; that is left to gardener-node-agent.
					Expect(script).NotTo(ContainSubstring("kubelet.service"))
					Expect(script).NotTo(ContainSubstring("no-content.service"))
					Expect(script).NotTo(ContainSubstring("LimitMEMLOCK"))

					// gardener-node-agent's cached state is dropped so it re-applies the full OSC onto the wiped /etc.
					Expect(script).To(ContainSubstring("rm -f " + nodeagentconfigv1alpha1.LastAppliedOperatingSystemConfigFilePath))
					Expect(script).To(ContainSubstring("rm -f " + nodeagentconfigv1alpha1.BaseDir + "/last-computed-osc-changes.yaml"))
				})

				It("should bind hook paths to gardener core's exported constants", func() {
					_, _, files, _, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					var hookFile *extensionsv1alpha1.File
					for i := range files {
						if files[i].Path == gardenlinux.PathEtcSetupHook {
							hookFile = &files[i]
							break
						}
					}
					Expect(hookFile).NotTo(BeNil())
					decoded, err := base64.StdEncoding.DecodeString(hookFile.Content.Inline.Data)
					Expect(err).NotTo(HaveOccurred())
					script := string(decoded)

					Expect(nodeagentconfigv1alpha1.UnitName).To(Equal("gardener-node-agent.service"))
					Expect(nodeagentconfigv1alpha1.LastAppliedOperatingSystemConfigFilePath).To(Equal("/var/lib/gardener-node-agent/last-applied-osc.yaml"))
					Expect(nodeagentconfigv1alpha1.BaseDir).To(Equal("/var/lib/gardener-node-agent"))
					Expect(script).To(ContainSubstring(nodeagentconfigv1alpha1.BaseDir + "/last-computed-osc-changes.yaml"))
				})

				It("should set 0755 permissions on the hook file", func() {
					_, _, files, _, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					for _, f := range files {
						if f.Path == gardenlinux.PathEtcSetupHook {
							Expect(f.Permissions).NotTo(BeNil())
							Expect(*f.Permissions).To(Equal(int32(0755)))
							return
						}
					}
					Fail("hook file not found in extension files")
				})
			})

			It("should add one empty additional unit for containerd", func() {
				_, units, files, _, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())
				Expect(units).To(HaveLen(1))
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

type multiPart struct {
	contentType string
	params      map[string]string
	content     string
}

func readMimeMultiParts(s string) []multiPart {
	GinkgoHelper()
	const (
		contentTypeIdentifier = "Content-Type"
		boundary              = "==BOUNDARY=="
	)

	var parts []multiPart

	mr := multipart.NewReader(strings.NewReader(s), boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		Expect(err).ShouldNot(HaveOccurred())

		c, err := io.ReadAll(p)
		Expect(err).ShouldNot(HaveOccurred())

		mediaType, params, err := mime.ParseMediaType(p.Header.Get(contentTypeIdentifier))
		Expect(err).ShouldNot(HaveOccurred())

		parts = append(parts, multiPart{
			contentType: mediaType,
			params:      params,
			content:     string(c),
		})
	}
	return parts
}

func extractVsmpConfiguration(p multiPart) map[string]string {
	GinkgoHelper()
	Expect(p.contentType).To(Equal("text/x-vsmp"))
	Expect(p.params).To(HaveLen(1))
	Expect(p.params).To(HaveKeyWithValue("section", "vsmp"))

	lines := strings.Split(p.content, "\n")

	config := make(map[string]string, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		Expect(found).To(BeTrue())
		config[key] = value
	}

	return config
}

func extractUserdata(p multiPart) string {
	GinkgoHelper()
	Expect(p.contentType).To(Equal("text/x-shellscript"))
	Expect(p.params).To(BeEmpty())
	return p.content
}

func decodeVsmpUserData(s string) (map[string]string, string) {
	GinkgoHelper()
	prefix := `Content-Type: multipart/mixed; boundary="==BOUNDARY=="
MIME-Version: 1.0`

	Expect(strings.HasPrefix(s, prefix)).To(BeTrue())
	parts := readMimeMultiParts(s)
	Expect(parts).To(HaveLen(2))
	vsmpConfig := extractVsmpConfiguration(parts[0])
	userData := extractUserdata(parts[1])
	return vsmpConfig, userData
}

func encodeMemoryOneConfigurationIntoOsc(codec runtime.Codec, osc *extensionsv1alpha1.OperatingSystemConfig, moc *memoryonegardenlinuxv1alpha1.OperatingSystemConfiguration) error {
	GinkgoHelper()
	encoded, err := runtime.Encode(codec, moc)
	if err != nil {
		return err
	}
	osc.Spec.ProviderConfig = &runtime.RawExtension{
		Raw: encoded,
	}
	return nil
}
