// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/kubelet"
	oscutils "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/utils"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/gardenlinux"
	"github.com/gardener/gardener-extension-os-gardenlinux/pkg/webhook/operatingsystemconfig"
)

func TestOperatingSystemConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook OperatingSystemConfig Suite")
}

var (
	fciCodec           = oscutils.NewFileContentInlineCodec()
	kubeletConfigCodec = kubelet.NewConfigCodec(fciCodec)
	logger             = logr.Discard()
	ctx                = context.Background()
	mgr                = test.FakeManager{}

	kubeletConfigTemplate = kubeletconfigv1beta1.KubeletConfiguration{
		CgroupDriver: "cgroupfs",
	}

	criConfig = extensionsv1alpha1.CRIConfig{
		Name:         "containerd",
		CgroupDriver: ptr.To(extensionsv1alpha1.CgroupDriverCgroupfs),
	}

	oscTemplate = extensionsv1alpha1.OperatingSystemConfig{
		Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
			DefaultSpec: extensionsv1alpha1.DefaultSpec{
				Type: gardenlinux.OSTypeGardenLinux,
			},
			CRIConfig: &criConfig,
		},
	}
)

var _ = Describe("Ensurer", func() {
	var (
		ensurer       genericmutator.Ensurer
		kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration
	)

	BeforeEach(func() {
		kubeletConfig = kubeletConfigTemplate.DeepCopy()
		ensurer = operatingsystemconfig.NewEnsurer(mgr, logger)
	})

	It("Should replace the cgroup driver in a kubelet config to systemd", func() {
		Expect(ensurer.EnsureKubeletConfiguration(ctx, nil, nil, kubeletConfig, nil)).To(Succeed())
		Expect(kubeletConfig.CgroupDriver).To(Equal(operatingsystemconfig.KubeletCgroupDriverSystemd))
	})

	It("Should not replace the cgroup driver in a kubelet config if it was previously empty", func() {
		kubeletConfig.CgroupDriver = ""
		Expect(ensurer.EnsureKubeletConfiguration(ctx, nil, nil, kubeletConfig, nil)).To(Succeed())
		Expect(kubeletConfig.CgroupDriver).To(BeEmpty())
	})

	It("Should replace the containerd cgroup driver in the CRIConfiguration", func() {
		c := criConfig.DeepCopy()
		Expect(ensurer.EnsureCRIConfig(ctx, nil, c, nil)).To(Succeed())
		Expect(*c.CgroupDriver).To(Equal(extensionsv1alpha1.CgroupDriverSystemd))
	})
})

var _ = Describe("Mutator", func() {
	var (
		ensurer genericmutator.Ensurer
		mutator webhook.Mutator
	)

	BeforeEach(func() {
		ensurer = operatingsystemconfig.NewEnsurer(mgr, logger)
		fciCodec := oscutils.NewFileContentInlineCodec()
		mutator = genericmutator.NewMutator(
			mgr,
			ensurer,
			oscutils.NewUnitSerializer(),
			kubelet.NewConfigCodec(fciCodec),
			fciCodec,
			logger,
		)
	})

	Describe("#MutateGardenLinuxOSC", func() {
		var (
			osc extensionsv1alpha1.OperatingSystemConfig
		)

		BeforeEach(func() {
			kubeletConfig := kubeletConfigTemplate.DeepCopy()
			files, err := filesWithKkubletConfig(kubeletConfig)
			Expect(err).To(BeNil())

			osc = *oscTemplate.DeepCopy()
			osc.Spec.Files = files
		})

		It("should not mutate an OperatingSystemConfig not for Garden Linux", func() {
			osc.Spec.Type = "not-gardenlinux"

			Expect(mutator.Mutate(ctx, &osc, nil)).To(Succeed())

			mutatedKubeletConfig, err := extractKubeletConfigCgroupDriver(osc.Spec.Files)
			Expect(err).To(BeNil())

			Expect(mutatedKubeletConfig.CgroupDriver).NotTo(Equal(operatingsystemconfig.KubeletCgroupDriverSystemd))
			Expect(osc.Spec.CRIConfig.CgroupDriver).NotTo(Equal(extensionsv1alpha1.CgroupDriverSystemd))
		})
	})
})

func extractKubeletConfigCgroupDriver(oscFiles []extensionsv1alpha1.File) (*kubeletconfigv1beta1.KubeletConfiguration, error) {
	var kubeletConfigFCI *extensionsv1alpha1.FileContentInline
	for _, f := range oscFiles {
		if f.Path != v1beta1constants.OperatingSystemConfigFilePathKubeletConfig {
			continue
		}

		kubeletConfigFCI = f.Content.Inline
	}

	if kubeletConfigFCI == nil {
		return nil, fmt.Errorf("no kubeletconfig found inline")
	}

	kubeletConfig, err := kubeletConfigCodec.Decode(kubeletConfigFCI)
	if err != nil {
		return nil, err
	}

	return kubeletConfig, nil
}

func filesWithKkubletConfig(kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) ([]extensionsv1alpha1.File, error) {
	kubeletConfigFci, err := kubeletConfigCodec.Encode(kubeletConfig, "b64")
	if err != nil {
		return nil, err
	}

	return []extensionsv1alpha1.File{
		{
			Path:        v1beta1constants.OperatingSystemConfigFilePathKubeletConfig,
			Permissions: ptr.To(int32(0644)),
			Content: extensionsv1alpha1.FileContent{
				Inline: kubeletConfigFci,
			},
		},
	}, nil
}
