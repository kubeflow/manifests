#!/usr/bin/env bash
# This script aims at helping create a PR to update the manifests of the
# kubeflow/spark-operator repository.
# This script:
# 1. Checks out a new branch
# 2. Generates the manifests using Helm
# 3. Commits the changes
#
# Afterwards the developers can submit the PR to the kubeflow/manifests
# repository, based on that local branch
# It must be executed directly from its directory

# strict mode http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euxo pipefail
IFS=$'\n\t'

# You can use tags or commit hashes
SPARK_OPERATOR_VERSION=${SPARK_OPERATOR_VERSION:="2.1.0"}
SPARK_OPERATOR_HELM_CHART_REPOSITORY=${SPARK_OPERATOR_HELM_CHART_REPOSITORY:="https://kubeflow.github.io/spark-operator"}
DEV_MODE=${DEV_MODE:=false}
BRANCH=${BRANCH:=synchronize-spark-operator-manifests-${SPARK_OPERATOR_VERSION?}}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)

echo "Creating branch: ${BRANCH}"
if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

# Create the branch in the manifests repository
if ! git show-ref --verify --quiet refs/heads/$BRANCH; then
    git checkout -b $BRANCH
else
    echo "Branch $BRANCH already exists."
fi

echo "Generating manifests from Helm chart version ${SPARK_OPERATOR_VERSION}..."

# Generate the manifests using Helm
DST_DIR=$MANIFESTS_DIR/apps/spark/spark-operator/base
mkdir -p $DST_DIR
cd $DST_DIR

# Generate the manifests using Helm
helm template -n kubeflow --include-crds spark-operator spark-operator \
--set "spark.jobNamespaces={}" \
--set webhook.enable=true \
--set webhook.port=9443 \
--version ${SPARK_OPERATOR_VERSION} \
--repo ${SPARK_OPERATOR_HELM_CHART_REPOSITORY} > resources.yaml

echo "Successfully generated manifests."

echo "Updating README..."
# Use OS-compatible sed command
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS version
    sed -i '' 's/Spark Operator[^|]*|[^|]*apps\/spark\/spark-operator[^|]*|[^|]*[0-9]\.[0-9]\.[0-9]/Spark Operator	|	apps\/spark\/spark-operator	|	'"${SPARK_OPERATOR_VERSION}"'/g' "${MANIFESTS_DIR}/README.md"
else
    # Linux version
    sed -i 's/Spark Operator.*|.*apps\/spark\/spark-operator[^|]*|.*[0-9]\.[0-9]\.[0-9]/Spark Operator	|	apps\/spark\/spark-operator	|	'"${SPARK_OPERATOR_VERSION}"'/g' "${MANIFESTS_DIR}/README.md"
fi

echo "Committing the changes..."
cd $MANIFESTS_DIR
git add apps/spark
git add README.md
git add scripts/synchronize-spark-operator-manifests.sh
git commit -s -m "Update kubeflow/spark-operator manifests to ${SPARK_OPERATOR_VERSION}"

echo "Changes committed to branch ${BRANCH}. You can now push and create a PR."