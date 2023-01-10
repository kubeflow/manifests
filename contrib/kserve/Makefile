KSERVE_VERSION ?= 0.9.0

.PHONY: upgrade-kserve-manifests
upgrade-kserve-manifests: clean-kserve-manifests
	curl -sSL 'https://github.com/kserve/kserve/releases/download/v$(KSERVE_VERSION)/kserve_kubeflow.yaml' -o 'kserve/kserve_kubeflow.yaml'
	curl -sSL 'https://github.com/kserve/kserve/releases/download/v$(KSERVE_VERSION)/kserve-runtimes.yaml' -o 'kserve/kserve-runtimes.yaml'
	curl -sSL 'https://github.com/kserve/kserve/releases/download/v$(KSERVE_VERSION)/kserve.yaml' -o 'kserve/kserve.yaml'

.PHONY: clean-kserve-manifests
clean-kserve-manifests:
	rm -f ./kserve/kserve*.yaml

.PHONY: install-kserve
install-kserve:
	kubectl apply -k kserve


