#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# kubeflow/model-registry repo.
# This script:
# 1. Checks out a new branch
# 2. Copies files to the correct places
# 3. Commits the changes
#
# Afterwards the developers can submit the PR to the kubeflow/manifests
# repo, based on that local branch

# strict mode http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

DEV_MODE=${DEV_MODE:=false}
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-model-registry}
BRANCH=${BRANCH:=sync-kubeflow-model-registry-manifests-${COMMIT?}}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)

if [ "$DEV_MODE" != "false" ]; then
    echo "WARNING: Dev mode enabled..."
fi

echo "Creating branch: ${BRANCH}"

# DEV: Comment out this if you are testing locally
if [ -n "$(git status --porcelain)" ]; then
  # Uncommitted changes
  echo "WARNING: You have uncommitted changes, exiting..."
  exit 1
fi

if [ `git branch --list $BRANCH` ]
then
   echo "WARNING: Branch $BRANCH already exists. Exiting..."
   exit 1
fi

# DEV: If you are testing locally set DEV_MODE=true to skip this step
if [ "$DEV_MODE" = "false" ]; then
    git checkout -b $BRANCH
fi

echo "Checking out in $SRC_DIR to $COMMIT..."
cd $SRC_DIR
if [ -n "$(git status --porcelain)" ]; then
  # Uncommitted changes
  echo "WARNING: You have uncommitted changes, exiting..."
  exit 1
fi
git checkout $COMMIT

echo "Copying model-registry manifests..."
DST_DIR=$MANIFESTS_DIR/apps/model-registry/upstream
rm -rf $DST_DIR
mkdir -p $DST_DIR
cp $SRC_DIR/manifests/kustomize/* $DST_DIR -r

echo "Successfully copied all manifests."

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/model-registry/tree/.*/manifests/kustomize)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/model-registry/tree/$COMMIT/manifests/kustomize)"

sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

# DEV: If you are testing locally set DEV_MODE=true to skip this step
if [ "$DEV_MODE" = "false" ]; then
    echo "Committing the changes..."
    cd $MANIFESTS_DIR
    git add apps
    git add README.md
    git commit -s -m "Update kubeflow/model-registry manifests from ${COMMIT}"
fi