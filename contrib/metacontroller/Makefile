BUILD_DIR=./build

.PHONY: hydrate
hydrate:
	rm -r -f "${BUILD_DIR}"
	mkdir -p "${BUILD_DIR}"
	kustomize build -o "${BUILD_DIR}" ./base
	echo "Succesfully built into ${BUILD_DIR}."

.PHONY: clean-build
clean-build:
	rm -r -f "${BUILD_DIR}"
	echo "Succesfully cleaned ${BUILD_DIR}."

.PHONY: apply
apply: hydrate
	kubectl apply -f "${BUILD_DIR}"
	echo "Succesfully applied from ${BUILD_DIR}."

.PHONY: delete
delete: hydrate
	kubectl delete -f "${BUILD_DIR}"
	echo "Succesfully deleted metacontroller."

.PHONY: test
test: test-deploy test-catset-controller test-deploy-cleanup

.PHONY: test-deploy
test-deploy:
	kubectl create namespace kubeflow || echo "Found 'kubeflow' namespace"
	kustomize build ./base | kubectl apply -f -
	sleep 30
	# Check CRDs
	kubectl -n kubeflow wait --for=condition=established --timeout=60s crd/compositecontrollers.metacontroller.k8s.io
	kubectl -n kubeflow wait --for=condition=established --timeout=60s crd/controllerrevisions.metacontroller.k8s.io
	kubectl -n kubeflow wait --for=condition=established --timeout=60s crd/decoratorcontrollers.metacontroller.k8s.io
	# Check ServiceAccount
	kubectl get -n kubeflow serviceaccount/meta-controller-service
	kubectl get -n kubeflow clusterrolebinding/meta-controller-cluster-role-binding
	# Check StatefulSet
	kubectl -n kubeflow rollout status --watch --timeout=60s StatefulSet/metacontroller
	echo "[Test] contrib/metacontroller deployment PASSED"

.PHONY: test-deploy-cleanup
test-deploy-cleanup:
	# Cleanup
	kustomize build ./base | kubectl delete -f -

.PHONY: test-catset-controller
test-catset-controller:
	# Create a sample CRD
	kubectl apply -f test/catset-controller.yaml
	sleep 30
	# Check a custom CRD
	kubectl -n kubeflow get compositecontrollers/catset-controller
	# Cleanup
	kubectl delete compositecontrollers/catset-controller
	echo "[Test] contrib/metacontroller catset-controller PASSED"

.PHONY: pull
pull:
	# Check ./UPGRADE.md for instructions on how to pull updated manifests
	kpt pkg update .@${KFP_VERSION}
