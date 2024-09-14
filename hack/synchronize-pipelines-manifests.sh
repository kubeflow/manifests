#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# kubeflow/pipelines repository.
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

COMMIT="2.3.0" # You can use tags as well
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-pipelines}
BRANCH=${BRANCH:=synchronize-kubeflow-pipelines-manifests-${COMMIT?}}

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

# Checkout the KFP repository
mkdir -p $SRC_DIR
cd $SRC_DIR
if [ ! -d "pipelines/.git" ]; then
    git clone https://github.com/kubeflow/pipelines.git
fi
cd $SRC_DIR/pipelines
if ! git rev-parse --verify --quiet $COMMIT; then
    git checkout -b $COMMIT
else
    git checkout $COMMIT
fi


if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

echo "Copying pipelines manifests..."
DST_DIR=$MANIFESTS_DIR/apps/pipeline/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
cp $SRC_DIR/pipelines/manifests/kustomize $DST_DIR -r


echo "Successfully copied all manifests."

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/pipelines/tree/.*/manifests/kustomize)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/pipelines/tree/$COMMIT/manifests/kustomize)"

sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Committing the changes..."
cd $MANIFESTS_DIR
git add apps
git add README.md
git commit -s -m "Update kubeflow/pipelines manifests from ${COMMIT}"
