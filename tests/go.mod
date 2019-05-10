module github.com/kubeflow/manifests

go 1.12

require (
	k8s.io/api/v2 v2.0.0
	k8s.io/apiextensions-apiserver/v2 v2.0.0
	k8s.io/apimachinery/v2 v2.0.0
	k8s.io/client-go/v2 v2.0.0
	sigs.k8s.io/kustomize/v2 v2.0.0-00010101000000-000000000000
)

replace (
	k8s.io/api/v2 => /tmp/v2/k8s.io/api
	k8s.io/apiextensions-apiserver/v2 => /tmp/v2/k8s.io/apiextensions-apiserver
	k8s.io/apimachinery/v2 => /tmp/v2/k8s.io/apimachinery
	k8s.io/client-go/v2 => /tmp/v2/k8s.io/client-go
	sigs.k8s.io/kustomize/v2 => /tmp/v2/sigs.k8s.io/kustomize
)
