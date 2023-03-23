#!/bin/bash

set -eux

kubectl get pods -A

mkdir stdout-${JOB_INDEX}

for NAMESPACE in kubeflow-user-example-com kubeflow; do
    for POD in $(kubectl get pods -n $NAMESPACE -o custom-columns=:metadata.name --no-headers); do
        for CONTAINER in $(kubectl get pods -n $NAMESPACE -o jsonpath="{.spec.containers[*].name}" $POD); do
          kubectl logs -n $NAMESPACE --timestamps $POD -c $CONTAINER > stdout-${JOB_INDEX}/$NAMESPACE-$POD-$CONTAINER.log || true
        done
    done
done

tar -cvzf stdout.tar.gz stdout-${JOB_INDEX}/*.log