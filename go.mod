module github.com/gardener/gardener-extension-os-gardenlinux

go 1.16

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/gardener/gardener v1.45.0
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.18.0
	github.com/spf13/cobra v1.2.1
	golang.org/x/tools v0.1.9
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.23.3
	k8s.io/component-base v0.23.5
	sigs.k8s.io/controller-runtime v0.11.2
)

replace (
	k8s.io/api => k8s.io/api v0.23.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.3
	k8s.io/autoscaler => k8s.io/autoscaler v0.0.0-20201008123815-1d78814026aa // translates to k8s.io/autoscaler/vertical-pod-autoscaler@v0.9.0
	k8s.io/autoscaler/vertical-pod-autoscaler => k8s.io/autoscaler/vertical-pod-autoscaler v0.9.0
	k8s.io/client-go => k8s.io/client-go v0.23.3
	k8s.io/code-generator => k8s.io/code-generator v0.23.3
	k8s.io/component-base => k8s.io/component-base v0.23.3
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
)

// workaround for https://github.com/gardener/hvpa-controller/issues/92, remove once it's fixed
replace (
	github.com/gardener/hvpa-controller => github.com/gardener/hvpa-controller v0.4.0
	github.com/gardener/hvpa-controller/api => github.com/gardener/hvpa-controller/api v0.4.0
)
