module github.com/yuyicai/kubei

go 1.13

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/fatih/color v1.7.0
	github.com/go-kratos/kratos v0.5.0
	github.com/heroku/docker-registry-client v0.0.0-20190909225348-afc9e1acc3d5
	github.com/lithammer/dedent v1.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/vbauerster/mpb/v6 v6.0.2
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.18.5
	sigs.k8s.io/yaml v1.2.0
)

replace (
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/apiextensions-apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/cli-runtime => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/client-go => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/cloud-provider => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/cluster-bootstrap => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/code-generator => k8s.io/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/component-base => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/cri-api => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/csi-translation-lib => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/kube-aggregator => k8s.io/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/kube-controller-manager => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/kube-proxy => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/kube-scheduler => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/kubectl => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/kubelet => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/legacy-cloud-providers => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/metrics => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/sample-apiserver => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/sample-cli-plugin => k8s.io/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20200626033830-e6503f8d8f76
	k8s.io/sample-controller => k8s.io/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20200626033830-e6503f8d8f76
)
