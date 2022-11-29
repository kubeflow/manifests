
SELDON_VERSION ?= 1.15.0

seldon-core-operator/base: clean-kustomize
	mkdir -p seldon-core-operator/base
	cd seldon-core-operator/base && helm template -f ../../values.yaml seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --namespace kubeflow --version ${SELDON_VERSION}  > resources.yaml
	sed -i 's#cert-manager.io/inject-ca-from:.*#cert-manager.io/inject-ca-from: kubeflow/seldon-serving-cert#g' seldon-core-operator/base/resources.yaml
	sed -i "s#'seldon-webhook-service.kubeflow.svc.cluster.local'#"'seldon-webhook-service.$$(SERVICE_NAMESPACE).svc.cluster.local#' seldon-core-operator/base/resources.yaml
	sed -i "s#commonName: 'seldon-webhook-service.kubeflow.svc'#"'commonName: seldon-webhook-service.$$(CERTIFICATE_NAMESPACE).svc#' seldon-core-operator/base/resources.yaml
	sed -i "s#'seldon-webhook-service.kubeflow.svc'#"'seldon-webhook-service.$$(SERVICE_NAMESPACE).svc#' seldon-core-operator/base/resources.yaml
	sed -i 's#kubeflow/kubeflow-gateway#$$(SERVICE_NAMESPACE)/kubeflow-gateway#' seldon-core-operator/base/resources.yaml
	cp kustomization.tpl seldon-core-operator/base/kustomization.yaml
	cp kustomizeconfig.tpl seldon-core-operator/base/kustomizeconfig.yaml

.PHONY:clean-kustomize
clean-kustomize:
	rm -rf seldon-core-operator/base

test:
	# Create seldon namespace
	kubectl create namespace seldon || echo "seldon namespace exists"
	# Create kubeflow namespace if not exists
	kubectl create namespace kubeflow || echo "kubeflow namespace exists"
	# Creating resources
	kustomize build seldon-core-operator/base | kubectl apply -n kubeflow -f -
	# Waiting for resources
	kubectl wait --for condition=Available -n kubeflow deployment/seldon-controller-manager
	# Wait for manager resources to be created
	sleep 5
	# Create example deployment
	kubectl apply -n seldon -f example.yaml
	# Wait for manager resources to be created
	sleep 5
	kubectl get deploy -A
	sleep 5
	kubectl logs -n kubeflow deployment/seldon-controller-manager
	sleep 5
	kubectl get deploy -A
	# Wait for deployment
	kubectl wait --for condition=Available -n seldon deployment/echo-default-0-classifier --timeout=360s
	# Remove deployment
	kubectl delete -n seldon -f example.yaml
	# Cleaning resources
	kustomize build seldon-core-operator/base | kubectl delete -n kubeflow -f -
