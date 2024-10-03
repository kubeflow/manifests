#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# contrib/seldon repository.
# This script:
# 1. Checks out a new branch
# 2. Copies files to the correct places
# 3. Commits the changes
#
# Afterwards the developers can submit the PR to the kubeflow/manifests
# repository, based on that local branch
# It must be executed directly from its directory

# strict mode http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euxo pipefail
IFS=$'\n\t'

COMMIT="v1.18.1" # You can use tags as well
SRC_DIR=${SRC_DIR:=/tmp/seldon}
BRANCH=${BRANCH:=synchronize-seldon-core-manifests-${COMMIT?}}
UPDATE_ECHO_MODEL=false

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)

echo "Creating branch: ${BRANCH}"

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi
if [ `git branch --list $BRANCH` ]
then
   echo "WARNING: Branch $BRANCH already exists."
fi

# Create the branch in the manifests repository
if ! git show-ref --verify --quiet refs/heads/$BRANCH; then
    git checkout -b $BRANCH
else
    echo "Branch $BRANCH already exists."
fi
echo "Checking out in $SRC_DIR to $COMMIT..."

# Checkout the Seldon repository
mkdir -p $SRC_DIR
cd $SRC_DIR
if [ ! -d "seldon-core/.git" ]; then
    git clone https://github.com/SeldonIO/seldon-core.git
fi
cd $SRC_DIR/seldon-core
if ! git rev-parse --verify --quiet $COMMIT; then
    git checkout -b $COMMIT
else
    git checkout $COMMIT
fi


if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

echo "Updating seldon manifests..."
DST_DIR=$MANIFESTS_DIR/contrib/seldon

cd $DST_DIR
SRC_TXT="SELDON_VERSION ?= .*"
DST_TXT="SELDON_VERSION ?= ${COMMIT:1}"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${DST_DIR}/Makefile
# Update manifests
SELDON_OPERATOR_CHART="$SRC_DIR/seldon-core/helm-charts/seldon-core-operator" make seldon-core-operator/base

echo "Successfully updated all manifests."

if [ "$UPDATE_ECHO_MODEL" = "true" ]; then
    echo "Updating seldonio/echo-model version..."
    SRC_TXT="seldonio/echo-model:[0-9]\+\.[0-9]\+\.[0-9]\+"
    DST_TXT="seldonio/echo-model:${COMMIT:1}"
    sed -i "s|$SRC_TXT|$DST_TXT|g" ${DST_DIR}/README.md
    sed -i "s|$SRC_TXT|$DST_TXT|g" ${DST_DIR}/example.yaml
    echo "Successfully updated seldonio/echo-model."
fi


echo "Committing the changes..."
cd $MANIFESTS_DIR
git add contrib/seldon
git add README.md
git commit -s -m "Update seldon manifests from ${COMMIT}"
