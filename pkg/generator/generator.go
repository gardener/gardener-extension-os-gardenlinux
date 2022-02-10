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
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux"
	gardenlinuxinstall "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux/install"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux/v1alpha1"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/version"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	cmd                = "/usr/bin/env bash %s"
	cloudInitGenerator *ostemplate.CloudInitGenerator
	client             runtimeclient.Client
	ctx                context.Context
	decoder            runtime.Decoder
)

//go:embed templates/*
var templates embed.FS

func init() {
	scheme := runtime.NewScheme()
	if err := gardenlinuxinstall.AddToScheme(scheme); err != nil {
		controllercmd.LogErrAndExit(err, "Could not update scheme")
	}
	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

	cloudInitTemplateString, err := templates.ReadFile(filepath.Join("templates", "cloud-init.gardenlinux.template"))
	runtimeutils.Must(err)

	cloudInitTemplate, err := ostemplate.NewTemplate("cloud-init").Parse(string(cloudInitTemplateString))
	runtimeutils.Must(err)

	cloudInitGenerator = ostemplate.NewCloudInitGenerator(cloudInitTemplate, ostemplate.DefaultUnitsPath, cmd, additionalValuesFunc)
}

// CloudInitGenerator is the generator which will genereta the cloud init yaml
func CloudInitGenerator(c context.Context) *ostemplate.CloudInitGenerator {
	ctx = c
	return cloudInitGenerator
}

func InjectClient(c runtimeclient.Client) {
	client = c
}

func additionalValuesFunc(osc *extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("seed client not initialized")
	}

	if osc.Spec.Type != gardenlinux.OSTypeGardenLinux {
		return nil, nil
	}

	obj := &v1alpha1.OperatingSystemConfiguration{}
	if osc.Spec.ProviderConfig != nil {
		if _, _, err := decoder.Decode(osc.Spec.ProviderConfig.Raw, nil, obj); err != nil {
			return nil, fmt.Errorf("failed to decode provider config: %+v", err)
		}
	}

	shoot, err := extensionscontroller.GetShoot(ctx, client, osc.Namespace)
	if err != nil {
		return nil, err
	}

	cGroupVersion, err := setCgroupVersion(obj.CgroupVersion, shoot.Spec.Kubernetes.Version)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{
		"LinuxSecurityModule": obj.LinuxSecurityModule,
		"NetFilterBackend":    obj.NetFilterBackend,
		"cGroupVersion":       cGroupVersion,
	}

	return values, nil
}

// setCgroupVersion determines the proper cGroup version depending on the Kubernetes version used in a shoot
func setCgroupVersion(cGroupVer v1alpha1.CgroupVersion, shootVersion string) (v1alpha1.CgroupVersion, error) {
	shootLessThan119, err := version.CompareVersions(shootVersion, "<", "1.19")
	if err != nil {
		return "", err
	}

	if shootLessThan119 {
		return v1alpha1.CgroupVersionV1, nil
	}

	return cGroupVer, nil
}
