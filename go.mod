module github.com/gardener/gardener-extension-os-gardenlinux

go 1.14

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/gardener/gardener v1.7.2
	github.com/gobuffalo/packr v1.30.1
	github.com/gobuffalo/packr/v2 v2.5.1
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/spf13/cobra v0.0.6
	k8s.io/apimachinery v0.17.6
	k8s.io/code-generator v0.17.6
	sigs.k8s.io/controller-runtime v0.5.5
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
	k8s.io/component-base => k8s.io/component-base v0.17.6
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
)
