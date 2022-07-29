#!/bin/bash
set -e
kubectl create ns argo
# Download the binary
curl -sLO https://github.com/argoproj/argo-workflows/releases/download/v3.1.3/argo-linux-amd64.gz
# Unzip
gunzip argo-linux-amd64.gz
# Make binary executable
chmod +x argo-linux-amd64
# Move binary to path
sudo mv ./argo-darwin-amd64 /usr/local/bin/argo