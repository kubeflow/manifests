#!/bin/bash

set -eux

kubectl describe pods -A

mkdir describe-${JOB_INDEX}

for RESOURCE in $(kubectl api-resources -o name | sort); do
  kubectl describe $RESOURCE -A > "describe-${JOB_INDEX}/$RESOURCE.describe" || true
done

tar -cvzf describe.tar.gz describe-${JOB_INDEX}/*.describe