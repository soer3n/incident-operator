module github.com/soer3n/incident-operator

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/prometheus/common v0.26.0
	github.com/soer3n/yaho v0.0.0-20211008165841-68d56bc1f45a
	github.com/spf13/cobra v1.1.3
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.6-0.20210908190839-cf92b39a962c // indirect
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v1.5.2
	k8s.io/kubectl v0.21.0
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/descheduler v0.19.0
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.0
)
