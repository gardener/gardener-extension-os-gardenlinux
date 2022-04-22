// Copyright (c) 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package gardenlinux

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OSTypeGardenLinux is a constant for the Garden Linux extension OS type.
const OSTypeGardenLinux = "gardenlinux"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatingSystemConfiguration allows to specify configuration for the operating system.
type OperatingSystemConfiguration struct {
	metav1.TypeMeta

	// LinuxSecurityModule allows to configure default Linux Security Module for Garden Linux. If not present, it will default to `AppArmor`.
	LinuxSecurityModule LinuxSecurityModule

	// NetFilterBackend allows to configure the netfilter backend to be used on Garden Linux
	NetFilterBackend NetFilterBackend

	// CgroupVersion allows to configure which cgroup version will be used on Garden Linux
	CgroupVersion CgroupVersion
}

// LinuxSecurityModule defines the Linux Security Module (LSM) for Garden Linux
type LinuxSecurityModule string

const (
	// LsmAppArmor is the identifier for AppArmor as LSM
	LsmAppArmor LinuxSecurityModule = "AppArmor"
	// LsmSeLinux is the identifier for SELinux as LSM
	LsmSeLinux LinuxSecurityModule = "SELinux"
)

// NetFilterBackend defines the netfilter backend for Garden Linux
type NetFilterBackend string

const (
	// NetFilterNfTables is the identifier for nftables as netfilter backend
	NetFilterNfTables NetFilterBackend = "nftables"
	// NetFilterIpTables is the identifier for nftables as netfilter backend
	NetFilterIpTables NetFilterBackend = "iptables"
)

// CgroupVersion defines the cgroup version (v1 or v2) to be configured on Garden Linux
type CgroupVersion string

const (
	// CgroupVersionV1 sets the cgroup version to (legacy) v1
	CgroupVersionV1 CgroupVersion = "v1"
	// CgroupVersionV2 sets the cgroup version to v2
	CgroupVersionV2 CgroupVersion = "v2"
)
