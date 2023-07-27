KUBERAY_RELEASE_VERSION ?= 0.4.0
KUBERAY_HELM_CHART_REPO ?= https://ray-project.github.io/kuberay-helm/

.PHONY: kuberay-operator/base
kuberay-operator/base:
	mkdir -p kuberay-operator/base
	cd kuberay-operator/base && helm template --include-crds kuberay-operator kuberay-operator --version ${KUBERAY_RELEASE_VERSION} --repo ${KUBERAY_HELM_CHART_REPO} > resources.yaml

.PHONY: test
test:
	./test.sh

