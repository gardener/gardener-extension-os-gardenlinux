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
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator/testfiles"
	"github.com/go-logr/logr"

	commongen "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator/test"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type byteSlice []byte

func (b byteSlice) GomegaString() string {
	delimiter := []byte("\n----------------------------\n")
	return string(append(b, delimiter...))
}

var (
	logger = logr.Discard()

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

	criConfigContainerd = extensionsv1alpha1.CRIConfig{
		Name: extensionsv1alpha1.CRINameContainerD,
	}

	osctemplate = commongen.OperatingSystemConfig{
		Object: &extensionsv1alpha1.OperatingSystemConfig{
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type: gardenlinux.OSTypeGardenLinux,
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

	osc commongen.OperatingSystemConfig
)

var _ = Describe("Garden Linux OS Generator Test", func() {

	Describe("Conformance Tests Bootstrap", func() {
		g := gardenlinux_generator.CloudInitGenerator()
		test.DescribeTest(gardenlinux_generator.CloudInitGenerator(), testfiles.Files)()

		BeforeEach(func() {
			osc = osctemplate
			osc.Bootstrap = true
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
		})

		It("[docker] [bootstrap] should render correctly", func() {
			e, err := testfiles.Files.ReadFile("docker-bootstrap")
			expectedCloudInit := byteSlice(e)
			Expect(err).NotTo(HaveOccurred())

			osc.CRI = nil
			c, _, err := g.Generate(logger, &osc)
			cloudInit := byteSlice(c)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[containerd] [bootstrap] should render correctly", func() {
			e, err := testfiles.Files.ReadFile("containerd-bootstrap")
			expectedCloudInit := byteSlice(e)
			Expect(err).NotTo(HaveOccurred())

			osc.CRI = &criConfigContainerd
			c, _, err := g.Generate(logger, &osc)
			cloudInit := byteSlice(c)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})
	})

	Describe("Conformance Tests Reconcile", func() {
		var g = gardenlinux_generator.CloudInitGenerator()

		BeforeEach(func() {
			osc = osctemplate
			osc.Bootstrap = false
			osc.Object.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
		})

		It("[docker] [reconcile] should render correctly", func() {
			e, err := testfiles.Files.ReadFile("docker-reconcile")
			Expect(err).NotTo(HaveOccurred())
			expectedCloudInit := byteSlice(e)

			c, _, err := g.Generate(logger, &osc)
			cloudInit := byteSlice(c)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})

		It("[containerd] [reconcile] should render correctly", func() {
			e, err := testfiles.Files.ReadFile("containerd-reconcile")
			Expect(err).NotTo(HaveOccurred())
			expectedCloudInit := byteSlice(e)

			osc.CRI = &criConfigContainerd
			c, _, err := g.Generate(logger, &osc)
			cloudInit := byteSlice(c)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloudInit).To(Equal(expectedCloudInit))
		})
	})
})
