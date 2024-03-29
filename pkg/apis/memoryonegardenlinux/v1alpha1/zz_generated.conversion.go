//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	memoryonegardenlinux "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/memoryonegardenlinux"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*OperatingSystemConfiguration)(nil), (*memoryonegardenlinux.OperatingSystemConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_OperatingSystemConfiguration_To_memoryonegardenlinux_OperatingSystemConfiguration(a.(*OperatingSystemConfiguration), b.(*memoryonegardenlinux.OperatingSystemConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*memoryonegardenlinux.OperatingSystemConfiguration)(nil), (*OperatingSystemConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_memoryonegardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(a.(*memoryonegardenlinux.OperatingSystemConfiguration), b.(*OperatingSystemConfiguration), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_OperatingSystemConfiguration_To_memoryonegardenlinux_OperatingSystemConfiguration(in *OperatingSystemConfiguration, out *memoryonegardenlinux.OperatingSystemConfiguration, s conversion.Scope) error {
	out.MemoryTopology = (*string)(unsafe.Pointer(in.MemoryTopology))
	out.SystemMemory = (*string)(unsafe.Pointer(in.SystemMemory))
	return nil
}

// Convert_v1alpha1_OperatingSystemConfiguration_To_memoryonegardenlinux_OperatingSystemConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_OperatingSystemConfiguration_To_memoryonegardenlinux_OperatingSystemConfiguration(in *OperatingSystemConfiguration, out *memoryonegardenlinux.OperatingSystemConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_OperatingSystemConfiguration_To_memoryonegardenlinux_OperatingSystemConfiguration(in, out, s)
}

func autoConvert_memoryonegardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(in *memoryonegardenlinux.OperatingSystemConfiguration, out *OperatingSystemConfiguration, s conversion.Scope) error {
	out.MemoryTopology = (*string)(unsafe.Pointer(in.MemoryTopology))
	out.SystemMemory = (*string)(unsafe.Pointer(in.SystemMemory))
	return nil
}

// Convert_memoryonegardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration is an autogenerated conversion function.
func Convert_memoryonegardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(in *memoryonegardenlinux.OperatingSystemConfiguration, out *OperatingSystemConfiguration, s conversion.Scope) error {
	return autoConvert_memoryonegardenlinux_OperatingSystemConfiguration_To_v1alpha1_OperatingSystemConfiguration(in, out, s)
}
