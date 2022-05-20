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
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux"
	gardenlinuxinstall "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux/install"

	gardenlinuxgenerator "github.com/gardener/gardener-extension-os-gardenlinux/pkg/generator/gardenlinux"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	cmd                = "/usr/bin/env bash %s"
	cloudInitGenerator *GardenLinuxCloudInitGenerator
	client             runtimeclient.Client
	ctx                context.Context
	decoder            runtime.Decoder
	unitsToEnable      []string
)

//go:embed templates/*
var templates embed.FS

// GardenLinuxCloudInitGenerator is a wrapper around the CloudInitGenerator from oscommon that implements its own generate function
type GardenLinuxCloudInitGenerator struct {
	cloudInitGenerator *ostemplate.CloudInitGenerator
}

func init() {
	scheme := runtime.NewScheme()
	if err := gardenlinuxinstall.AddToScheme(scheme); err != nil {
		runtimelog.Log.Error(err, "Could not update scheme")
		os.Exit(1)
	}

	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

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
func (g *GardenLinuxCloudInitGenerator) Generate(osc *generator.OperatingSystemConfig) ([]byte, *string, error) {

	if osc.Object.Spec.Type == gardenlinux.OSTypeGardenLinux &&
		osc.Object.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile {

		// get additional cgroupUnit and add them to the osc object before calling the generator
		cgroupScript, cgroupUnit, err := gardenlinuxgenerator.ConfigureCgroups(ctx, client, osc.Object, decoder)
		if err != nil {
			return nil, nil, err
		}

		// get settings for Linux Security Modules and netfilter backend
		lsmScript, lsmUnit, err := gardenlinuxgenerator.ConfigureLinuxSecurityModule(osc.Object, decoder)
		if err != nil {
			return nil, nil, err
		}

		// place the systemd unit and script that can restart the system if required
		restartScript, restartUnit, err := gardenlinuxgenerator.PlaceRestartUnit()
		if err != nil {
			return nil, nil, err
		}

		osc.Files = append(osc.Files, cgroupScript, lsmScript, restartScript)
		osc.Units = append(osc.Units, cgroupUnit, lsmUnit, restartUnit)
		unitsToEnable = []string{cgroupUnit.Name, lsmUnit.Name, restartUnit.Name}

		// add scripts and dropins for containerd and kubelet
		runtimeScripts, runtimeUnits, err := gardenlinuxgenerator.ContainerRuntimesCgroupDrivers()
		if err != nil {
			return nil, nil, err
		}

		osc.Files = append(osc.Files, runtimeScripts...)
		osc.Units = append(osc.Units, runtimeUnits...)
	}

	return g.cloudInitGenerator.Generate(osc)
}

// CloudInitGenerator is the generator which will genereta the cloud init yaml
func CloudInitGenerator(c context.Context) *GardenLinuxCloudInitGenerator {
	ctx = c
	return cloudInitGenerator
}

// InjectClient injects a client to the seed into the extension
func InjectClient(c runtimeclient.Client) {
	client = c
}
