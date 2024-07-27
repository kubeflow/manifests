#!/bin/bash
set -e
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.0/manifests/tigera-operator.yaml
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.0/manifests/custom-resources.yaml

kubectl wait --for=condition=Ready pods --all --namespace=calico-system --timeout 300s