SPARK_OPERATOR_RELEASE_VERSION ?= 2.0.1
SPARK_OPERATOR_HELM_CHART_REPO ?= https://kubeflow.github.io/spark-operator

.PHONY: spark-operator/base
spark-operator/base:
	mkdir -p spark-operator/base
	cd spark-operator/base && helm template --include-crds spark-operator spark-operator --version ${SPARK_OPERATOR_RELEASE_VERSION} --repo ${SPARK_OPERATOR_HELM_CHART_REPO} > resources.yaml

.PHONY: test
test:
	./test.sh ${KF_PROFILE}