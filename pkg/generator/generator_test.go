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
	gardenlinux_generator "github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator/testfiles"
	"github.com/go-logr/logr"

	commongen "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator/test"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var logger = logr.Discard()

var (
	permissions = int32(0644)
	unit1       = "unit1"
	unit2       = "unit2"
	ccdService  = "cloud-config-downloader.service"
	dropin      = "dropin"
	filePath    = "/var/lib/kubelet/ca.crt"

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

	criConfig = v1alpha1.CRIConfig{
		Name: v1alpha1.CRINameContainerD,
	}

	osc = commongen.OperatingSystemConfig{
		Object: &v1alpha1.OperatingSystemConfig{
			Spec: v1alpha1.OperatingSystemConfigSpec{
				Purpose: v1alpha1.OperatingSystemConfigPurposeProvision,
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
		g := gardenlinux_generator.CloudInitGenerator()
		test.DescribeTest(gardenlinux_generator.CloudInitGenerator(), testfiles.Files)()

		It("[docker] [bootstrap] should render correctly ", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("docker-bootstrap")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = true
			osc.Object.Spec.Purpose = v1alpha1.OperatingSystemConfigPurposeProvision
			osc.CRI = nil
			cloudInit, _, err := g.Generate(logger, &osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[docker] [reconcile] should render correctly", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("docker-reconcile")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = false
			osc.Object.Spec.Purpose = v1alpha1.OperatingSystemConfigPurposeReconcile
			osc.CRI = nil
			cloudInit, _, err := g.Generate(logger, &osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[containerd] [bootstrap] should render correctly", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("containerd-bootstrap")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = true
			osc.Object.Spec.Purpose = v1alpha1.OperatingSystemConfigPurposeProvision
			osc.CRI = &criConfig
			cloudInit, _, err := g.Generate(logger, &osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[containerd] [reconcile] should render correctly bootstrap", func() {
			expectedCloudInit, err := testfiles.Files.ReadFile("containerd-reconcile")
			Expect(err).NotTo(HaveOccurred())

			osc.Bootstrap = false
			osc.Object.Spec.Purpose = v1alpha1.OperatingSystemConfigPurposeReconcile
			osc.CRI = &criConfig
			cloudInit, _, err := g.Generate(logger, &osc)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})
	})
})
