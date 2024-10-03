#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# kubeflow/training-operator repository.
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

COMMIT="v1.8.1" # You can use tags as well
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-training-operator}
BRANCH=${BRANCH:=synchronize-kubeflow-training-operator-manifests-${COMMIT?}}

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

# Checkout the Training Operator repository
mkdir -p $SRC_DIR
cd $SRC_DIR
if [ ! -d "training-operator/.git" ]; then
    git clone https://github.com/kubeflow/training-operator.git
fi
cd $SRC_DIR/training-operator
if ! git rev-parse --verify --quiet $COMMIT; then
    git checkout -b $COMMIT
else
    git checkout $COMMIT
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

echo "Copying training-operator manifests..."
DST_DIR=$MANIFESTS_DIR/apps/training-operator/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
cp $SRC_DIR/training-operator/manifests $DST_DIR -r


echo "Successfully copied all manifests."

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/training-operator/tree/.*/manifests)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/training-operator/tree/$COMMIT/manifests)"

sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

# DEV: Comment out these commands if you are testing locally
echo "Committing the changes..."
cd $MANIFESTS_DIR
git add apps
git add README.md
git commit -s -m "Update kubeflow/training-operator manifests from ${COMMIT}"
