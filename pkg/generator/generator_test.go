// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package generator_test

import (
	"context"
	"encoding/json"

	gardenlinux_core "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux"
	gardenlinux_v1alpha1 "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux/v1alpha1"
	gardenlinux_generator "github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator/testfiles"

	commongen "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator/test"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx         context.Context = context.Background()
	permissions                 = int32(0644)
	unit1                       = "unit1"
	unit2                       = "unit2"
	ccdService                  = "cloud-config-downloader.service"
	dropin                      = "dropin"
	filePath                    = "/var/lib/kubelet/ca.crt"

	ctrl           *gomock.Controller
	c              *mockclient.MockClient
	clusterKey     client.ObjectKey
	shoot          *gardencorev1beta1.Shoot
	cluster        *extensionsv1alpha1.Cluster
	providerConfig *gardenlinux_v1alpha1.OperatingSystemConfiguration

	unitContent = []byte(`[Unit]
Description=test content
[Install]
WantedBy=multi-user.target
[Service]
Restart=always`)
	dropinContent = []byte(`[Service]
Environment="DOCKER_OPTS=--log-opt max-size=60m --log-opt max-file=3"`)

	fileContent = []byte(`secretRef:
name: default-token-d9nzl
dataKey: token`)

	criConfig = extensionsv1alpha1.CRIConfig{
		Name: extensionsv1alpha1.CRINameContainerD,
	}

	osc = commongen.OperatingSystemConfig{
		Object: &extensionsv1alpha1.OperatingSystemConfig{
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type: gardenlinux_core.OSTypeGardenLinux,
				},
			},
		},
		Units: []*commongen.Unit{
			{
				Name:    unit1,
				Content: unitContent,
			},
			{
				Name:    unit2,
				Content: unitContent,
				DropIns: []*commongen.DropIn{
					{
						Name:    dropin,
						Content: dropinContent,
					},
				},
			},
			{
				Name: ccdService,
			},
		},
		Files: []*commongen.File{
			{
				Path:        filePath,
				Content:     fileContent,
				Permissions: &permissions,
			},
		},
	}
)

var _ = Describe("Garden Linux OS Generator Test", func() {

	Describe("Conformance Tests", func() {
		g := gardenlinux_generator.CloudInitGenerator(ctx)
		test.DescribeTest(gardenlinux_generator.CloudInitGenerator(ctx), testfiles.Files)()

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			c = mockclient.NewMockClient(ctrl)
			gardenlinux_generator.InjectClient(c)

			c.EXPECT().Get(ctx, clusterKey, gomock.AssignableToTypeOf(&extensionsv1alpha1.Cluster{})).DoAndReturn(
				func(_ context.Context, namespacedname client.ObjectKey, obj *extensionsv1alpha1.Cluster) error {
					*obj = *cluster
					obj.ObjectMeta.Namespace = clusterKey.Namespace
					obj.ObjectMeta.Name = clusterKey.Name
					return nil
				})

			shoot = &gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{
					Kubernetes: gardencorev1beta1.Kubernetes{
						Version: "v1.19.0",
					},
				},
			}

			cluster = &extensionsv1alpha1.Cluster{
				Spec: extensionsv1alpha1.ClusterSpec{
					Shoot: runtime.RawExtension{
						Raw: encode(shoot),
					},
				},
			}

			providerConfig = &gardenlinux_v1alpha1.OperatingSystemConfiguration{
				LinuxSecurityModule: gardenlinux_v1alpha1.LsmAppArmor,
				NetFilterBackend:    gardenlinux_v1alpha1.NetFilterNfTables,
				CgroupVersion:       gardenlinux_v1alpha1.CgroupVersionV2,
			}

			osc.Object.Spec.ProviderConfig = &runtime.RawExtension{
				Raw: encode(providerConfig),
			}
		})

		It("[docker] [bootstrap] should render correctly", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("docker-bootstrap")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = true
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
			osc.CRI = nil
			cloudInit, _, err := g.Generate(&osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[docker] [reconcile] should render correctly", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("docker-reconcile")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = false
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
			osc.CRI = nil
			cloudInit, _, err := g.Generate(&osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[containerd] [bootstrap] should render correctly", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("containerd-bootstrap")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = true
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
			osc.CRI = &criConfig
			cloudInit, _, err := g.Generate(&osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[containerd] [reconcile] should render correctly bootstrap", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("containerd-reconcile")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = false
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
			osc.CRI = &criConfig
			cloudInit, _, err := g.Generate(&osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[bootstrap] should set the cgroup version to v1 for K8S <= 1.18", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("settings-cgroup-k8sv118")
			Expect(err).NotTo(HaveOccurred())

			shoot.Spec.Kubernetes.Version = "v1.18.0"
			cluster = &extensionsv1alpha1.Cluster{
				Spec: extensionsv1alpha1.ClusterSpec{
					Shoot: runtime.RawExtension{
						Raw: encode(shoot),
					},
				},
			}

			osc.Bootstrap = true
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
			osc.CRI = nil
			cloudInit, _, err := g.Generate(&osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
