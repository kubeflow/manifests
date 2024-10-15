SPARK_OPERATOR_RELEASE_VERSION ?= 2.0.2
SPARK_OPERATOR_HELM_CHART_REPO ?= https://kubeflow.github.io/spark-operator

.PHONY: spark-operator/base
spark-operator/base:
	mkdir -p spark-operator/base
	cd spark-operator/base && \
	helm template -n kubeflow --include-crds spark-operator spark-operator \
	--set "spark.jobNamespaces={}" \
	--set webhook.enable=true \
    --set webhook.port=9443 \
	--version ${SPARK_OPERATOR_RELEASE_VERSION} \
	--repo ${SPARK_OPERATOR_HELM_CHART_REPO} > resources.yaml

.PHONY: test
test:
	./test.sh ${KF_PROFILE}
