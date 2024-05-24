// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig

import (
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/kubelet"
	oscutils "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
)

const (
	// WebhookName is the webhook name.
	WebhookName = "os-gardenlinux-webhook"
	// WebhookPath is the webhook path.
	WebhookPath = "/webhooks/gardenlinux-osc"
)

var logger = log.Log.WithName(WebhookName)

// AddToManager returns a new mutating webhook that changes an OperatingSystemConfig for Garden Linux
func AddToManager(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	logger.Info("Adding webhook to manager")

	fciCodec := oscutils.NewFileContentInlineCodec()

	mutator := genericmutator.NewMutator(
		mgr,
		NewEnsurer(mgr, logger),
		oscutils.NewUnitSerializer(),
		kubelet.NewConfigCodec(fciCodec),
		fciCodec,
		logger,
	)

	objTypes := []extensionswebhook.Type{
		{Obj: &extensionsv1alpha1.OperatingSystemConfig{}},
	}

	handler, err := extensionswebhook.NewBuilder(mgr, logger).WithPredicates(isGardenLinuxOsc()).WithMutator(mutator, objTypes...).Build()
	if err != nil {
		return nil, err
	}

	webhook := &extensionswebhook.Webhook{
		Name:     extensionswebhook.PrefixedName(WebhookName),
		Provider: "",
		Action:   extensionswebhook.ActionMutating,
		Path:     WebhookPath,
		Target:   extensionswebhook.TargetSeed,
		Webhook:  &admission.Webhook{Handler: handler},
		Types:    objTypes,
	}

	return webhook, nil
}

// isGardenLinuxOsc returns a predicate that filters OperatingSystemConfigs just for Garden Linux
func isGardenLinuxOsc() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		osc, ok := obj.(*extensionsv1alpha1.OperatingSystemConfig)
		if !ok {
			return false
		}
		return osc.Spec.Type == gardenlinux.OSTypeGardenLinux
	})
}
