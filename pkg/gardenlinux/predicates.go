// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gardenlinux

import (
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// isGardenLinuxOsc returns a predicate that filters OperatingSystemConfigs just for Garden Linux
func PredicateGardenLinuxOsc() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		osc, ok := obj.(*extensionsv1alpha1.OperatingSystemConfig)
		if !ok {
			return false
		}
		return osc.Spec.Type == OSTypeGardenLinux || osc.Spec.Type == OSTypeGardenLinuxFips
	})
}
