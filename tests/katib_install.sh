#!/bin/bash
set -euxo pipefail

sudo apt-get update
sudo apt-get install -y apparmor-profiles
sudo apparmor_parser -R /etc/apparmor.d/usr.sbin.mysqld

cd applications/katib/upstream && kustomize build installs/katib-with-kubeflow | kubectl apply -f - && cd ../../../

kubectl wait --for=condition=Available deployment/katib-controller -n kubeflow --timeout=300s

kubectl wait --for=condition=Available deployment/katib-mysql -n kubeflow --timeout=300s

kubectl label namespace $KF_PROFILE katib.kubeflow.org/metrics-collector-injection=enabled --overwrite
