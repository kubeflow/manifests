BENTOML_YATAI_IMAGE_BUILDER_VERSION ?= 1.1.3
BENTOML_YATAI_DEPLOYMENT_VERSION ?= 1.1.4
BENTOML_HELM_CHART_REPO ?= https://bentoml.github.io/helm-charts

.PHONY: bentoml-yatai-stack/bases
bentoml-yatai-stack/bases: clean-kustomize
	mkdir -p bentoml-yatai-stack/bases/yatai-image-builder
	cd bentoml-yatai-stack/bases/yatai-image-builder && helm template --skip-tests yatai-image-builder-crds yatai-image-builder-crds --repo ${BENTOML_HELM_CHART_REPO} --namespace kubeflow --version ${BENTOML_YATAI_IMAGE_BUILDER_VERSION} > resources.yaml
	cd bentoml-yatai-stack/bases/yatai-image-builder && helm template --skip-tests -f ../../../sources/yatai-image-builder-values.yaml yatai-image-builder yatai-image-builder --repo ${BENTOML_HELM_CHART_REPO} --namespace kubeflow --version ${BENTOML_YATAI_IMAGE_BUILDER_VERSION} >> resources.yaml
	cp sources/kustomization-template.yaml bentoml-yatai-stack/bases/yatai-image-builder/kustomization.yaml
	mkdir -p bentoml-yatai-stack/bases/yatai-deployment
	cd bentoml-yatai-stack/bases/yatai-deployment && helm template --skip-tests yatai-deployment-crds yatai-deployment-crds --repo ${BENTOML_HELM_CHART_REPO} --namespace kubeflow --version ${BENTOML_YATAI_DEPLOYMENT_VERSION} > resources.yaml
	cd bentoml-yatai-stack/bases/yatai-deployment && helm template --skip-tests -f ../../../sources/yatai-deployment-values.yaml yatai-deployment yatai-deployment --repo ${BENTOML_HELM_CHART_REPO} --namespace kubeflow --version ${BENTOML_YATAI_DEPLOYMENT_VERSION} >> resources.yaml
	cp sources/kustomization-template.yaml bentoml-yatai-stack/bases/yatai-deployment/kustomization.yaml

.PHONY: clean-kustomize
clean-kustomize:
	rm -rf bentoml-yatai-stack/bases

.PHONY: test
test:
	./test.sh

