#!/usr/bin/env bash
# This script helps to create a PR to update the manifests
set -euxo pipefail
IFS=$'\n\t'

KSERVE_VERSION="v0.14.1"
COMMIT="v0.14.1" # You can use tags as well
SRC_DIR=${SRC_DIR:=/tmp/kserve}
BRANCH=${BRANCH:=synchronize-kserve-manifests-${COMMIT?}}

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

# Checkout the kserve repository
mkdir -p $SRC_DIR
cd $SRC_DIR
if [ ! -d "kserve/.git" ]; then
    git clone https://github.com/kserve/kserve.git
fi
cd $SRC_DIR/kserve
if ! git rev-parse --verify --quiet $COMMIT; then
    git checkout -b $COMMIT
else
    git checkout $COMMIT
fi


if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

echo "Copying kserve manifests..."
DST_DIR=$MANIFESTS_DIR/apps/kserve/kserve
if [ -d "$DST_DIR" ]; then
    rm -rf "$DST_DIR"/kserve*
fi
cp $SRC_DIR/kserve/install/"$KSERVE_VERSION"/* $DST_DIR -r


echo "Successfully copied all manifests."

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/kserve/kserve/releases/tag/.*)"
DST_TXT="\[$COMMIT\](https://github.com/kserve/kserve/releases/tag/$COMMIT/install/$KSERVE_VERSION)"

sed -i "s|$SRC_TXT|$DST_TXT|g" "${MANIFESTS_DIR}"/README.md

echo "Committing the changes..."
cd "$MANIFESTS_DIR"
git add apps/kserve
git add README.md
git commit -s -m "Update kserve manifests from ${KSERVE_VERSION}" -m "Update kserve/kserve manifests from ${COMMIT}"
