// Copyright (c) 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_OperatingSystemConfiguration sets the defaults for the Garden Linux operating system configuration
func SetDefaults_OperatingSystemConfiguration(obj *OperatingSystemConfiguration) {
	if isEmptyString(obj.MemoryTopology) {
		obj.MemoryTopology = pointer.String("2")
	}

	if isEmptyString(obj.SystemMemory) {
		obj.SystemMemory = pointer.String("6x")
	}
}

// isEmptyString returns true if pointer to a string is either nil or the string has zero length
func isEmptyString(s *string) bool {
	if s == nil || len(*s) == 0 {
		return true
	}
	return false
}
