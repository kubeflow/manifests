
seldon-core:
	git clone -b v1.0.2 --single-branch git@github.com:SeldonIO/seldon-core.git

seldon-core-operator/base: seldon-core clean-kustomize
	mkdir -p seldon-core-operator/base
	helm template -f values.yaml seldon-core/helm-charts/seldon-core-operator --namespace kubeflow | sed 's/1.0.2-SNAPSHOT/1.0.2/g' > seldon-core-operator/base/resources.yaml
	cp kustomization.tpl seldon-core-operator/base/kustomization.yaml

.PHONY: clean
clean:
	rm -rf seldon-core

.PHONY:clean-kustomize
clean-kustomize:
	rm -rf seldon-core-operator/base
