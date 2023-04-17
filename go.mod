module github.com/softonic/pod-defaulter

go 1.14

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/ghodss/yaml v1.0.0
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/imdario/mergo v0.3.9
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	golang.org/x/crypto v0.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/autoscaler/vertical-pod-autoscaler v0.0.0-20200910092546-63259fb5dd89
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.8
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	knative.dev/pkg v0.0.0-20200911235400-de640e81d149
	knative.dev/test-infra v0.0.0-20200911201000-3f90e7c8f2fa
	sigs.k8s.io/controller-tools v0.11.3 // indirect
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.3

	k8s.io/client-go => k8s.io/client-go v0.18.3
	k8s.io/code-generator => k8s.io/code-generator v0.18.3
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)
