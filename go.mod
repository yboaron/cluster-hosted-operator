module github.com/yboaron/cluster-hosted-operator

go 1.13

require (
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/golangci/golangci-lint v1.32.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/openshift/api v0.0.0-20200827090112-c05698d102cf
	github.com/openshift/client-go v0.0.0-20200827190008-3062137373b5
	github.com/openshift/cluster-network-operator v0.0.0-20201105033330-1ee0aaf1bdb8
	github.com/openshift/library-go v0.0.0-20201013192036-5bd7c282e3e7
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/controller-tools v0.3.0
	sigs.k8s.io/kustomize/kustomize/v3 v3.8.5

)

replace k8s.io/client-go => k8s.io/client-go v0.19.0
