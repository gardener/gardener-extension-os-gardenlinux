// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"embed"
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/controller/operatingsystemconfig/generator/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/memoryone"
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
func additionalValues(osc *extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error) {
	values := map[string]interface{}{
		"unitsToEnable": unitsToEnable,
	}

	if err := memoryone.MemoryOneValues(osc, values); err != nil {
		return nil, err
	}

	return values, nil
}

// Generate generates a Garden Linux specific cloud-init script from the given OperatingSystemConfig.
func (g *GardenLinuxCloudInitGenerator) Generate(logger logr.Logger, osc *generator.OperatingSystemConfig) ([]byte, *string, error) {
	// we are only setting this up if the worker pool is configured with containerd
	if osc.Object.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile &&
		osc.CRI != nil && osc.CRI.Name == extensionsv1alpha1.CRINameContainerD {

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
		containerdCgroupdriverScript, containerdCgroupdriverDropin, err := gardenlinux.ContainerdCgroupDriver()
		if err != nil {
			return nil, nil, err
		}

		osc.Files = append(osc.Files, containerdCgroupdriverScript...)
		osc.Units = append(osc.Units, containerdCgroupdriverDropin...)

	}

	return g.cloudInitGenerator.Generate(logger, osc)
}

// CloudInitGenerator is the generator which will genereta the cloud init yaml
func CloudInitGenerator() *GardenLinuxCloudInitGenerator {
	return cloudInitGenerator
}
