KSERVE_VERSION ?= 0.10.0
MODELS_WEBAPP_VERSION ?= 0.8.1
MODELS_WEBAPP_RELEASE_VERSION := $(shell echo ${MODELS_WEBAPP_VERSION} | cut -d "." -f1-2)

.PHONY: upgrade-kserve-manifests
upgrade-kserve-manifests: clean-kserve-manifests
	curl -sSL 'https://github.com/kserve/kserve/releases/download/v$(KSERVE_VERSION)/kserve_kubeflow.yaml' -o 'kserve/kserve_kubeflow.yaml'
	curl -sSL 'https://github.com/kserve/kserve/releases/download/v$(KSERVE_VERSION)/kserve-runtimes.yaml' -o 'kserve/kserve-runtimes.yaml'
	curl -sSL 'https://github.com/kserve/kserve/releases/download/v$(KSERVE_VERSION)/kserve.yaml' -o 'kserve/kserve.yaml'

.PHONY: clean-kserve-manifests
clean-kserve-manifests:
	cd kserve && rm -f kserve.yaml kserve-runtimes.yaml kserve_kubeflow.yaml

.PHONY: install-kserve
install-kserve:
	# Create kubeflow namespace if not exists
	kubectl create namespace kubeflow || echo "kubeflow namespace exists"
	kubectl apply -k kserve

.PHONY: uninstall-kserve
uninstall-kserve:
	kubectl delete -k kserve

.PHONY: test-kserve
test-kserve:
	cd tests && pytest .

.PHONY: clean-models-webapp-manifests
clean-models-webapp-manifests:
	rm -rf models-web-app

.PHONY: upgrade-models-webapp-manifests
upgrade-models-webapp-manifests: clean-models-webapp-manifests
	git clone https://github.com/kserve/models-web-app.git kserve-models-web-app-src
	cd kserve-models-web-app-src && git checkout release-${MODELS_WEBAPP_RELEASE_VERSION}
	cp -r kserve-models-web-app-src/config ./models-web-app
	# cleanup
	rm -rf kserve-models-web-app-src

.PHONY: install-models-web-app
install-models-webapp:
	# Create kubeflow namespace if not exists
	kubectl create namespace kubeflow || echo "kubeflow namespace exists"
	kubectl apply -k models-web-app/overlays/kubeflow

.PHONY: uninstall-models-web-app
uninstall-models-webapp:
	kubectl delete -k models-web-app/overlays/kubeflow

.PHONY: test-models-webapp
test-models-webapp: install-models-webapp
	kubectl wait --for=condition=Available --timeout=300s -n kubeflow deployment/kserve-models-web-app
	@echo "Test Passed"
	@echo "cleaning up ..."
	kubectl delete -k models-web-app/overlays/kubeflow
