#!/bin/bash
#
# TODO(jlewi): This script is outdated. You probably want to use
# kubeflow/testing/py/kubeflow/testing/tools/applications.py
# see https://github.com/kubeflow/testing/pull/596

# Replace 'app.kubernetes.io/version: v0.7.x' with 'app.kubernetes.io/version: v1.0.0'
grep -rl --exclude-dir={kfdef,gatekeeper,gcp/deployment_manager_configs,aws/infra_configs,docs,hack,plugins} 'app.kubernetes.io/version: v0.7' ./ \
  | xargs sed -i -E 's/app.kubernetes.io\/version: v0.7(.*)/app.kubernetes.io\/version: v1.0.0/g'

# Replace 'app.kubernetes.io/instance: <application>-v0.7.x' with 'app.kubernetes.io/instance: <application>-v1.0.0'
grep -rl --exclude-dir={kfdef,gatekeeper,gcp/deployment_manager_configs,aws/infra_configs,docs,hack,plugins} 'app.kubernetes.io/instance: [a-z\-]*-v0.7' ./ \
  | xargs sed -i -E 's/app.kubernetes.io\/instance: (.+)-v0.7(.*)/app.kubernetes.io\/instance: \1-v1.0.0/g'
