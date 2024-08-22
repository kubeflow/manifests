
KUSTOMIZE = $(shell pwd)/bin/kustomize

# Define the target
.PHONY: install
install: kustomize ## Install Kubeflow into the K8s cluster specified in ~/.kube/config.
	@echo "Applying Kubernetes resources using Kustomize"
	@while ! $(KUSTOMIZE) build example | kubectl apply -f -; do \
		echo "Retrying to apply resources"; \
		sleep 20; \
	done


PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	GOBIN=$(PROJECT_DIR)/bin go install sigs.k8s.io/kustomize/kustomize/v5@v5.2.1
