module github.com/gardener/gardener-extension-runtime-gvisor

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.9.0 // indirect
	github.com/gardener/controller-manager-library v0.0.0-00010101000000-000000000000 // indirect
	github.com/gardener/gardener v1.2.1-0.20200402092110-3e4c4917c83f
	github.com/gardener/gardener-extensions v1.5.1-0.20200402130253-88d7a59e5b63
	github.com/gardener/gardener-resource-manager v0.10.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr/v2 v2.1.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/mock v1.3.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.17.0
	k8s.io/component-base v0.17.0
	k8s.io/helm v2.16.1+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	github.com/gardener/controller-manager-library => github.com/gardener/controller-manager-library v0.1.1-0.20191212112146-917449ad760c
	github.com/gardener/gardener => github.com/danielfoehrKn/gardener v0.0.0-20200403101853-662170f9c149
	github.com/gardener/gardener-extensions => github.com/gardener/gardener-extensions v1.5.1-0.20200402130253-88d7a59e5b63
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f // kubernetes-1.16.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 // kubernetes-1.16.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad // kubernetes-1.16.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // kubernetes-1.16.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3 // kubernetes-1.16.0
)
