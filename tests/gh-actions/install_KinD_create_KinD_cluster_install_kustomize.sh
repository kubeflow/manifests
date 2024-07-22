#!/bin/bash

set -e

# Function to handle errors
handle_error() {
    echo "Error: $1"
    exit 1
}

# Install KinD
echo "Installing KinD..."
if ! ./tests/gh-actions/install_kind.sh; then
    handle_error "Failed to install KinD"
fi

# Create KinD Cluster
echo "Creating KinD Cluster..."
if ! kind create cluster --config tests/gh-actions/kind-cluster.yaml; then
    handle_error "Failed to create KinD cluster"
fi

# Install kustomize
echo "Installing kustomize..."
if ! ./tests/gh-actions/install_kustomize.sh; then
    handle_error "Failed to install kustomize"
fi

echo "All steps completed successfully."
