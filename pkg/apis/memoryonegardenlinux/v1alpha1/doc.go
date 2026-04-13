// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// +k8s:deepcopy-gen=package
// +k8s:conversion-gen=github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/memoryonegardenlinux
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta

//go:generate crd-ref-docs --source-path=. --config=../../../../hack/api-reference/memoryonegardenlinux-config.yaml --renderer=markdown --templates-dir=$GARDENER_HACK_DIR/api-reference/template --log-level=ERROR --output-path=../../../../hack/api-reference/memoryonegardenlinux.md

// Package v1alpha1 contains the v1alpha1 version of the API.
// +groupName=memoryone-gardenlinux.os.extensions.gardener.cloud
package v1alpha1 // import "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/memoryonegardenlinux/v1alpha1"
