#!/bin/bash
set -e
curl --silent --location --remote-name "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.0.3/kustomize_v5.0.3_linux_amd64.tar.gz"
tar -xzvf kustomize_v5.0.3_linux_amd64.tar.gz
chmod a+x kustomize
sudo mv kustomize /usr/local/bin/kustomize
