// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_OperatingSystemConfiguration sets the defaults for the Garden Linux operating system configuration
func SetDefaults_OperatingSystemConfiguration(obj *OperatingSystemConfiguration) {
	if isEmptyString(obj.MemoryTopology) {
		obj.MemoryTopology = ptr.To("2")
	}

	if isEmptyString(obj.SystemMemory) {
		obj.SystemMemory = ptr.To("6x")
	}
}

// isEmptyString returns true if pointer to a string is either nil or the string has zero length
func isEmptyString(s *string) bool {
	if s == nil || len(*s) == 0 {
		return true
	}
	return false
}
