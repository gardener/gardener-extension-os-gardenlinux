//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright (c) SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	gardenlinux "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/gardenlinux"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*OperatingSystemConfiguration)(nil), (*gardenlinux.OperatingSystemConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_OperatingSystemConfiguration_To_gardenlinux_OperatingSystemConfiguration(a.(*OperatingSystemConfiguration), b.(*gardenlinux.OperatingSystemConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*gardenlinux.OperatingSystemConfiguration)(nil), (*OperatingSystemConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_gardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(a.(*gardenlinux.OperatingSystemConfiguration), b.(*OperatingSystemConfiguration), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_OperatingSystemConfiguration_To_gardenlinux_OperatingSystemConfiguration(in *OperatingSystemConfiguration, out *gardenlinux.OperatingSystemConfiguration, s conversion.Scope) error {
	out.LinuxSecurityModule = gardenlinux.LinuxSecurityModule(in.LinuxSecurityModule)
	out.NetFilterBackend = gardenlinux.NetFilterBackend(in.NetFilterBackend)
	out.CgroupVersion = gardenlinux.CgroupVersion(in.CgroupVersion)
	return nil
}

// Convert_v1alpha1_OperatingSystemConfiguration_To_gardenlinux_OperatingSystemConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_OperatingSystemConfiguration_To_gardenlinux_OperatingSystemConfiguration(in *OperatingSystemConfiguration, out *gardenlinux.OperatingSystemConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_OperatingSystemConfiguration_To_gardenlinux_OperatingSystemConfiguration(in, out, s)
}

func autoConvert_gardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(in *gardenlinux.OperatingSystemConfiguration, out *OperatingSystemConfiguration, s conversion.Scope) error {
	out.LinuxSecurityModule = LinuxSecurityModule(in.LinuxSecurityModule)
	out.NetFilterBackend = NetFilterBackend(in.NetFilterBackend)
	out.CgroupVersion = CgroupVersion(in.CgroupVersion)
	return nil
}

// Convert_gardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration is an autogenerated conversion function.
func Convert_gardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(in *gardenlinux.OperatingSystemConfiguration, out *OperatingSystemConfiguration, s conversion.Scope) error {
	return autoConvert_gardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(in, out, s)
}
