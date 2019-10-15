#!/bin/bash

# Replace 'app.kubernetes.io/version: v0.6.x' with 'app.kubernetes.io/version: v0.7.0'
grep -rl --exclude-dir=hack 'app.kubernetes.io/version: v0.6' ./ | xargs sed -i -E 's/app.kubernetes.io\/version: v0.6(.*)/app.kubernetes.io\/version: v0.7.0/g'

# Replace 'app.kubernetes.io/instance: <application>-v0.6.x' with 'app.kubernetes.io/instance: <application>-v0.7.0'
grep -rl --exclude-dir=hack 'app.kubernetes.io/instance: [a-z\-]*-v0.6' ./ | xargs sed -i -E 's/app.kubernetes.io\/instance: (.+)-v0.6(.*)/app.kubernetes.io\/instance: \1-v0.7.0/g'
