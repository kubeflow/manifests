#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# kserve/models-web-app repository.
# This script:
# 1. Checks out a new branch
# 2. Copies files to the correct places
# 3. Commits the changes
#
# Afterwards the developers can submit the PR to the kubeflow/manifests
# repository, based on that local branch


COMMIT="0.13.0" # You can use tags as well
SRC_DIR=${SRC_DIR:=/tmp/kserve-models-web-app}
BRANCH=${BRANCH:=synchronize-kserve-web-app-manifests-${COMMIT?}}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)

echo "Creating branch: ${BRANCH}"

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

if [ "$(git branch --list $BRANCH)" ]
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

# Checkout the Model Registry repository
mkdir -p $SRC_DIR
cd $SRC_DIR || exit
if [ ! -d "models-web-app/.git" ]; then
    git clone https://github.com/kserve/models-web-app.git
fi
cd $SRC_DIR/models-web-app || exit
if ! git rev-parse --verify --quiet $COMMIT; then
    git checkout -b $COMMIT
else
    git checkout $COMMIT
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

echo "Copying admission-webhook manifests..."
DST_DIR=$MANIFESTS_DIR/contrib/kserve/models-web-app
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp $SRC_DIR/models-web-app/config/* $DST_DIR -r

echo "Successfully copied all manifests."

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kserve/models-web-app/tree/.*)"
DST_TXT="\[$COMMIT\](https://github.com/kserve/models-web-app/tree/$COMMIT/config)"

sed -i "s|$SRC_TXT|$DST_TXT|g" "${MANIFESTS_DIR}"/README.md

echo "Committing the changes..."
cd $MANIFESTS_DIR || exit
git add contrib/kserve/models-web-app
git add README.md
git commit -s -m "Update kserve models web application manifests from ${COMMIT}"
