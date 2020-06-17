#!/usr/bin/env bash

export PIPELINES_SRC_REPO=https://github.com/kubeflow/pipelines.git
export PIPELINES_VERSION=1.0.0-rc.2
# Pulling for the first time
# kpt pkg get $PIPELINES_SRC_REPO/manifests/kustomize@$PIPELINES_VERSION pipeline/upstream

# Updates
kpt pkg update pipeline/upstream/@$PIPELINES_VERSION --strategy force-delete-replace
