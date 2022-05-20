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

package generator

import (
	"embed"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator/gardenlinux"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	cmd                = "/usr/bin/env bash %s"
	cloudInitGenerator *GardenLinuxCloudInitGenerator
	unitsToEnable      []string
)

//go:embed templates/*
var templates embed.FS

// GardenLinuxCloudInitGenerator is a wrapper around the CloudInitGenerator from oscommon that implements its own generate function
type GardenLinuxCloudInitGenerator struct {
	cloudInitGenerator *ostemplate.CloudInitGenerator
}

func init() {
	cloudInitTemplateString, err := templates.ReadFile(filepath.Join("templates", "cloud-init.gardenlinux.template"))
	runtimeutils.Must(err)

	cloudInitTemplate, err := ostemplate.NewTemplate("cloud-init").Parse(string(cloudInitTemplateString))
	runtimeutils.Must(err)

	cloudInitGenerator = &GardenLinuxCloudInitGenerator{
		cloudInitGenerator: ostemplate.NewCloudInitGenerator(cloudInitTemplate, ostemplate.DefaultUnitsPath, cmd, additionalValues),
	}
}

// additionalValues provides additional values to the cloud-init template
func additionalValues(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error) {
	return map[string]interface{}{
		"unitsToEnable": unitsToEnable,
	}, nil
}

// Generate generates a Garden Linux specific cloud-init script from the given OperatingSystemConfig.
func (g *GardenLinuxCloudInitGenerator) Generate(logger logr.Logger, osc *generator.OperatingSystemConfig) ([]byte, *string, error) {

	if osc.Object.Spec.Type == gardenlinux.OSTypeGardenLinux &&
		osc.Object.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile {

		// add additional scripts that are provided in the embedded fs
		additionalScripts, err := gardenlinux.GetAdditionalScripts()
		if err != nil {
			return nil, nil, err
		}
		osc.Files = append(osc.Files, additionalScripts...)

		// add scripts and dropins for kubelet
		kubeletCgroupdriverScript, kubeletCgroupdriverDropin, err := gardenlinux.KubeletCgroupDriver()
		if err != nil {
			return nil, nil, err
		}
		osc.Files = append(osc.Files, kubeletCgroupdriverScript...)
		osc.Units = append(osc.Units, kubeletCgroupdriverDropin...)

		// add scripts and dropins for containerd if activated
		if osc.CRI != nil && osc.CRI.Name == extensionsv1alpha1.CRINameContainerD {
			containerdCgroupdriverScript, containerdCgroupdriverDropin, err := gardenlinux.ContainerdCgroupDriver()
			if err != nil {
				return nil, nil, err
			}

			osc.Files = append(osc.Files, containerdCgroupdriverScript...)
			osc.Units = append(osc.Units, containerdCgroupdriverDropin...)
		}
	}

	return g.cloudInitGenerator.Generate(logger, osc)
}

// CloudInitGenerator is the generator which will genereta the cloud init yaml
func CloudInitGenerator() *GardenLinuxCloudInitGenerator {
	return cloudInitGenerator
}
