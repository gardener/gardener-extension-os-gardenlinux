// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package memoryonegardenlinux

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatingSystemConfiguration infrastructure configuration resource
type OperatingSystemConfiguration struct {
	metav1.TypeMeta

	// MemoryTopology allows to configure the `mem_topology` parameter. If not present, it will default to `2`.
	MemoryTopology *string
	// SystemMemory allows to configure the `system_memory` parameter. If not present, it will default to `6x`.
	SystemMemory *string
}
