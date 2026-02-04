// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig

import (
	"context"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	extensionscontextwebhook "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewEnsurer creates a new operatingsystemconfig ensurer.
func NewEnsurer(mgr manager.Manager, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		logger: logger.WithName(strings.Join([]string{WebhookName, "ensurer"}, "-")),
		client: mgr.GetClient(),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer

	client client.Client
	logger logr.Logger
}

const (
	KubeletCgroupDriverSystemd = "systemd"
)

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the desired specification
func (e *ensurer) EnsureKubeletConfiguration(_ context.Context, _ extensionscontextwebhook.GardenContext, _ *semver.Version, new, _ *kubeletconfigv1beta1.KubeletConfiguration) error {
	e.logger.Info("Ensuring Kubelet cgroup driver")
	return ensureKubeletUsesSystemdCgroupDriver(new)
}

// EnsureContainerdConfig ensures the CRI config.
func (e *ensurer) EnsureCRIConfig(_ context.Context, _ extensionscontextwebhook.GardenContext, new, _ *extensionsv1alpha1.CRIConfig) error {
	e.logger.Info("Ensuring containerd cgroup driver")
	return ensureContainerdUsesSystemdCgroupDriver(new)
}

// ensureKubeletUsesSystemdCgroupDriver ensures that the kubelet configuration contains systemd as cgroup driver
func ensureKubeletUsesSystemdCgroupDriver(kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) error {
	if len(kubeletConfig.CgroupDriver) != 0 {
		kubeletConfig.CgroupDriver = KubeletCgroupDriverSystemd
	}
	return nil
}

// ensureContainerdUsesSystemdCgroupDriver ensures that the CRI configuration contains systemd as cgroup driver
func ensureContainerdUsesSystemdCgroupDriver(containerdConfig *extensionsv1alpha1.CRIConfig) error {
	containerdConfig.CgroupDriver = ptr.To(extensionsv1alpha1.CgroupDriverSystemd)
	return nil
}
