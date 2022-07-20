#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of the
# kserve/kserve repo.
# This script:
# 1. Checks out a new branch
# 2. Copies files to the correct places
# 3. Commits the changes
#
# Afterwards the developers can submit the PR to the kubeflow/manifests
# repo, based on that local branch

# Run this script form the root of kubeflow/manifests repository
# ./hack/sync-kserve-manifests.sh

# strict mode http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

CLONE_DIR=${CLONE_DIR:=/tmp}
KSERVE_DIR="${CLONE_DIR?}/kserve"
BRANCH=${BRANCH:=sync-kserve-manifests-${KSERVE_COMMIT?}}
# *_VERSION vars are required only if COMMIT does not match a tag
KSERVE_VERSION=${KSERVE_VERSION:=${KSERVE_COMMIT?}}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
MANIFESTS_DIR=$(dirname "${SCRIPT_DIR}")

echo "Creating branch: ${BRANCH}"

# DEV: Comment out this if when local testing
if [ -n "$(git status --porcelain)" ]; then
  # Uncommitted changes
  echo "WARNING: You have uncommitted changes, exiting..."
  exit 1
fi

if [ "$(git branch --list "${BRANCH}")" ]
then
   echo "WARNING: Branch ${BRANCH} already exists. Exiting..."
   exit 1
fi

# DEV: Comment out this checkout command when local testing
git checkout -b "${BRANCH}"

echo "Checking out in $KSERVE_DIR to $KSERVE_COMMIT..."
pushd $CLONE_DIR
    if [ ! -d "$KSERVE_DIR" ]
    then
        git clone https://github.com/kserve/kserve.git && cd kserve
        git checkout "${KSERVE_COMMIT}"
    else
        echo "WARNING: ${KSERVE_DIR} directory already exists. Exiting..."
        exit 1
    fi
popd

echo "Copying kserve manifests..."
SRC_MANIFEST_PATH="$KSERVE_DIR"/install/"$KSERVE_VERSION"
if [ ! -d "$SRC_MANIFEST_PATH" ]
then
    echo "Directory $SRC_MANIFEST_PATH DOES NOT exists."
    exit 1
fi

DST_DIR=$MANIFESTS_DIR/contrib/kserve/kserve
pushd "$DST_DIR"
    rm -rf kserve*
popd
cp "$SRC_MANIFEST_PATH"/* "$DST_DIR" -r


echo "Successfully copied kserve manifests."

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kserve/kserve/tree/.*)"
DST_TXT="\[$KSERVE_COMMIT\](https://github.com/kserve/kserve/tree/$KSERVE_COMMIT/install/$KSERVE_VERSION)"

sed -i "s|$SRC_TXT|$DST_TXT|g" "${MANIFESTS_DIR}"/README.md

# DEV: Comment out these commands when local testing
echo "Committing the changes..."
cd "$MANIFESTS_DIR"
git add contrib/kserve
git add README.md
git commit -m "Update kserve manifests from ${KSERVE_VERSION}" -m "Update kserve/kserve manifests from ${KSERVE_COMMIT}"
