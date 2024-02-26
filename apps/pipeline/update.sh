#!/bin/bash
REPOSITORY_URL="https://github.com/kubeflow/pipelines"
TAG="2.0.5"
FOLDER_NAME="manifests/kustomize"
TMP_DIR=$(mktemp -d)
rm -rf ./upstream/

DOWNLOAD_URL="$REPOSITORY_URL/archive/refs/tags/${TAG}.zip"

curl -o "archive.zip" -L "$DOWNLOAD_URL"
unzip -oq "archive.zip"
cp -r "pipelines-$TAG/$FOLDER_NAME" ./upstream/
rm -rf "pipelines-$TAG"
rm -rf archive.zip
