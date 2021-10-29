module github.com/soer3n/incident-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/soer3n/yaho v0.0.0-20211008185703-d974e70a33a2
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.6-0.20210908190839-cf92b39a962c // indirect
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v1.5.2
	k8s.io/kubectl v0.21.0
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/descheduler v0.21.0
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/client-go => k8s.io/client-go v0.21.0
)
