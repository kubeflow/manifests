#!/bin/bash

# Replace 'app.kubernetes.io/version: v0.7.0' with 'app.kubernetes.io/version:v0.7.1'
grep -rl --exclude-dir={kfdef,gatekeeper,gcp/deployment_manager_configs,aws/infra_configs,docs,hack,plugins} 'app.kubernetes.io/version: v0.7.0' ./ \
  | xargs sed -i -E 's/app.kubernetes.io\/version: v0.7.0/app.kubernetes.io\/version: v0.7.1/g'

# Replace 'app.kubernetes.io/instance: <application>-v0.7.0' with 'app.kubernetes.io/instance: <application>-v0.7.1'
grep -rl --exclude-dir={kfdef,gatekeeper,gcp/deployment_manager_configs,aws/infra_configs,docs,hack,plugins} 'app.kubernetes.io/instance: [a-z\-]*-v0.7.0' ./ \
  | xargs sed -i -E 's/app.kubernetes.io\/instance: (.+)-v0.7.0/app.kubernetes.io\/instance: \1-v0.7.1/g'
