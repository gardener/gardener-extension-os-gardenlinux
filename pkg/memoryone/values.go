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

func MemoryOneValues(osc *extensionsv1alpha1.OperatingSystemConfig, values map[string]interface{}) error {
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
