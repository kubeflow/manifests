#!/bin/bash
set -euxo pipefail

KF_PROFILE="kubeflow-user-example-com"

kubectl apply -f tests/workspacekind.spark.test.yaml
kubectl apply -f tests/workspace.spark.test.kubeflow-user-example-com.yaml
kubectl wait --for=jsonpath='{.status.state}'=Running \
  workspace/spark-test -n "${KF_PROFILE}" \
  --timeout=600s

WORKSPACE_POD="$(kubectl -n "${KF_PROFILE}" get pods \
  -l notebooks.kubeflow.org/workspace-name=spark-test \
  -o jsonpath='{.items[0].metadata.name}')"

# The Workspace ServiceAccount is granted kubeflow-spark-edit by the WorkspaceKind,
# so it can create SparkConnect resources. Assert that before running the session,
# so an RBAC regression fails here with a clear message rather than inside Spark.
kubectl auth can-i create sparkconnects \
  --as="system:serviceaccount:${KF_PROFILE}:ws-spark-test" \
  -n "${KF_PROFILE}" | grep -qx yes

# The image ships the Kubeflow SDK but not its "spark" extra. Install the extra at the
# version already present rather than letting pip resolve to the latest release, so the
# test exercises the SDK a Workspace user actually has and needs no manual bump when the
# image moves. The resolver reports protobuf conflicts against the pre-installed kfp
# packages; they do not affect Spark.
SDK_VERSION="$(kubectl -n "${KF_PROFILE}" exec "${WORKSPACE_POD}" -- \
  python -c 'import kubeflow; print(kubeflow.__version__)')"
echo "Workspace image SDK version: ${SDK_VERSION}"

kubectl -n "${KF_PROFILE}" exec "${WORKSPACE_POD}" -- \
  python -m pip install --quiet "kubeflow[spark]==${SDK_VERSION}"

# Pin the Spark Connect client to the server version the SDK will provision. The spark
# extra and constants.DEFAULT_SPARK_VERSION can disagree -- in SDK 0.5.0 the extra pins
# pyspark-connect 4.2.0 while the constant is 4.0.4 -- so derive the client version from
# the constant rather than trusting the extra. If they diverge in the SDK the image ships,
# the test fails here, which is the right place to find out.
SPARK_VERSION="$(kubectl -n "${KF_PROFILE}" exec "${WORKSPACE_POD}" -- python -c \
  'from kubeflow.spark.backends.kubernetes import constants; print(constants.DEFAULT_SPARK_VERSION)')"
echo "SDK default Spark version: ${SPARK_VERSION}"

kubectl -n "${KF_PROFILE}" exec "${WORKSPACE_POD}" -- \
  python -m pip install --quiet "pyspark-connect==${SPARK_VERSION}"

kubectl -n "${KF_PROFILE}" cp \
  ./tests/spark_connect_from_workspace.py \
  "${WORKSPACE_POD}:/home/jovyan/spark_connect_from_workspace.py"

kubectl -n "${KF_PROFILE}" exec "${WORKSPACE_POD}" -- \
  env KF_PROFILE="${KF_PROFILE}" \
  python /home/jovyan/spark_connect_from_workspace.py