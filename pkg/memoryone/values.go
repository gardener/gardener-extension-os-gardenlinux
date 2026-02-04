// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package memoryone

import (
	"fmt"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"

	memoryonegardenlinux "github.com/gardener/gardener-extension-os-gardenlinux/pkg/apis/memoryonegardenlinux/v1alpha1"
)

var decoder runtime.Decoder

func init() {
	scheme := runtime.NewScheme()
	runtimeutils.Must(memoryonegardenlinux.AddToScheme(scheme))
	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
}

func Configuration(osc *extensionsv1alpha1.OperatingSystemConfig) (*memoryonegardenlinux.OperatingSystemConfiguration, error) {
	if osc.Spec.ProviderConfig == nil {
		return nil, nil
	}

	obj := &memoryonegardenlinux.OperatingSystemConfiguration{}
	if _, _, err := decoder.Decode(osc.Spec.ProviderConfig.Raw, nil, obj); err != nil {
		return nil, fmt.Errorf("failed to decode provider config: %+v", err)
	}

	return obj, nil
}

func Values(osc *extensionsv1alpha1.OperatingSystemConfig, values map[string]any) error {
	if osc.Spec.Type == OSTypeMemoryOneGardenLinux {
		config, err := Configuration(osc)
		if err != nil {
			return err
		}

		if config.MemoryTopology != nil {
			values["MemoryOneMemoryTopology"] = *config.MemoryTopology
		}

		if config.SystemMemory != nil {
			values["MemoryOneSystemMemory"] = *config.SystemMemory
		}
	}

	return nil
}
