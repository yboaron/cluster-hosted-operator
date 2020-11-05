module github.com/yboaron/cluster-hosted-operator

go 1.13

require (
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/openshift/cluster-network-operator v0.0.0-20201105033330-1ee0aaf1bdb8
	github.com/pkg/errors v0.9.1
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.3

)

replace k8s.io/client-go => k8s.io/client-go v0.19.0
