// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatingSystemConfiguration allows to specify configuration for the operating system.
type OperatingSystemConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// MemoryTopology allows to configure the `mem_topology` parameter. If not present, it will default to `2`.
	// +optional
	MemoryTopology *string `json:"memoryTopology,omitempty"`
	// SystemMemory allows to configure the `system_memory` parameter. If not present, it will default to `6x`.
	// +optional
	SystemMemory *string `json:"systemMemory,omitempty"`
}
