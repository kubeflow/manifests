#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# kubeflow/kubeflow repo.
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

COMMIT="v1.9.2" # You can use tags as well
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-kubeflow}
BRANCH=${BRANCH:=synchronize-kubeflow-kubeflow-manifests-${COMMIT?}}

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

# Checkout the upstream repository
mkdir -p $SRC_DIR
cd $SRC_DIR
if [ ! -d "kubeflow/.git" ]; then
    git clone https://github.com/kubeflow/kubeflow.git
fi
cd $SRC_DIR/kubeflow
if ! git rev-parse --verify --quiet $COMMIT; then
    git checkout -b $COMMIT
else
    git checkout $COMMIT
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

echo "Copying admission-webhook manifests..."
DST_DIR=$MANIFESTS_DIR/apps/admission-webhook/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/admission-webhook/manifests/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/admission-webhook/manifests)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/admission-webhook/manifests)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying centraldashboard manifests..."
DST_DIR=$MANIFESTS_DIR/apps/centraldashboard/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/centraldashboard/manifests/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/centraldashboard/manifests)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/centraldashboard/manifests)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying jupyter-web-app manifests..."
DST_DIR=$MANIFESTS_DIR/apps/jupyter/jupyter-web-app/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/crud-web-apps/jupyter/manifests/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/crud-web-apps/jupyter/manifests)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/crud-web-apps/jupyter/manifests)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying volumes-web-app manifests..."
DST_DIR=$MANIFESTS_DIR/apps/volumes-web-app/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/crud-web-apps/volumes/manifests/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/crud-web-apps/volumes/manifests)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/crud-web-apps/volumes/manifests)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying tensorboards-web-app manifests..."
DST_DIR=$MANIFESTS_DIR/apps/tensorboard/tensorboards-web-app/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/crud-web-apps/tensorboards/manifests/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/crud-web-apps/tensorboards/manifests)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/crud-web-apps/tensorboards/manifests)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying profile-controller manifests..."
DST_DIR=$MANIFESTS_DIR/apps/profiles/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/profile-controller/config/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/profile-controller/config)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/profile-controller/config)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying notebook-controller manifests..."
DST_DIR=$MANIFESTS_DIR/apps/jupyter/notebook-controller/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/notebook-controller/config/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/notebook-controller/config)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/notebook-controller/config)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying tensorboard-controller manifests..."
DST_DIR=$MANIFESTS_DIR/apps/tensorboard/tensorboard-controller/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/tensorboard-controller/config/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/tensorboard-controller/config)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/tensorboard-controller/config)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Copying pvcviewer-controller manifests..."
DST_DIR=$MANIFESTS_DIR/apps/pvcviewer-controller/upstream
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/kubeflow/components/pvcviewer-controller/config/* $DST_DIR -r

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kubeflow/kubeflow/tree/.*/components/pvcviewer-controller/config)"
DST_TXT="\[$COMMIT\](https://github.com/kubeflow/kubeflow/tree/$COMMIT/components/pvcviewer-controller/config)"
sed -i "s|$SRC_TXT|$DST_TXT|g" ${MANIFESTS_DIR}/README.md

echo "Successfully copied all manifests."

echo "Committing the changes..."
cd $MANIFESTS_DIR
git add apps
git add README.md
git commit -s -m "Update kubeflow/kubeflow manifests from ${COMMIT}"
