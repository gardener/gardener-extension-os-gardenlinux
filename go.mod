module github.com/gardener/gardener-extension-os-gardenlinux

go 1.16

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/gardener/gardener v1.24.1-0.20210609080620-7f7fbf24575c
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.5
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/code-generator v0.20.7
	k8s.io/component-base v0.20.7
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	k8s.io/api => k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.7
	k8s.io/client-go => k8s.io/client-go v0.20.7
	k8s.io/code-generator => k8s.io/code-generator v0.20.7
	k8s.io/component-base => k8s.io/component-base v0.20.7
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
)
