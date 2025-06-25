#!/bin/bash
set -e

HELM_VERSION="v3.16.4"

echo "Install Helm..."
{
    curl -Lo ./helm.tar.gz https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz
    tar -xzf helm.tar.gz linux-amd64/helm
    chmod +x linux-amd64/helm
    sudo mv linux-amd64/helm /usr/local/bin/helm
    rm -rf helm.tar.gz linux-amd64
} || { echo "Failed to install Helm"; exit 1; }

echo "Helm ${HELM_VERSION} installed successfully" 