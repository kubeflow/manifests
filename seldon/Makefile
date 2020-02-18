
seldon-core-operator/base: clean-kustomize
	mkdir -p seldon-core-operator/base
	cd seldon-core-operator/base && helm template -f ../../values.yaml seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --namespace kubeflow --version 1.0.2  > resources.yaml
	cp kustomization.tpl seldon-core-operator/base/kustomization.yaml

.PHONY:clean-kustomize
clean-kustomize:
	rm -rf seldon-core-operator/base
