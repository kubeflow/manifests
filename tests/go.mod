module github.com/kubeflow/manifests/tests

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/aws/aws-sdk-go v1.17.7
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190711103511-473e67f1d7d2 // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190711103511-473e67f1d7d2 // indirect
	github.com/fatih/color v1.7.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/jsonpointer v0.19.2 // indirect
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48
	github.com/google/go-cmp v0.3.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-getter v1.0.2
	github.com/imdario/mergo v0.3.7
	github.com/kubernetes-sigs/application v0.8.0
	github.com/onrik/logrus v0.2.1
	github.com/operator-framework/operator-sdk v0.13.0
	github.com/otiai10/copy v1.0.2
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	go.uber.org/zap v1.12.0 // indirect
	golang.org/dl v0.0.0-20191220003028-84da647cadd0 // indirect
	golang.org/x/crypto v0.0.0
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.10.0
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubernetes v1.16.2
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/v3 v3.1.0
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/go-openapi/jsonpointer => github.com/go-openapi/jsonpointer v0.17.0
	github.com/go-openapi/jsonreference => github.com/go-openapi/jsonreference v0.17.0
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.18.0
	github.com/go-openapi/swag => github.com/go-openapi/swag v0.17.0
	github.com/mitchellh/go-homedir => github.com/mitchellh/go-homedir v1.0.0
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2-0.20180428102519-11635eb403ff // indirect
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/kubernetes => k8s.io/kubernetes v1.15.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e
	sigs.k8s.io/application => sigs.k8s.io/application v0.0.0-20190404151855-67ae7f915d4e
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0
	sigs.k8s.io/kustomize/v3 => sigs.k8s.io/kustomize/v3 v3.1.0
)
