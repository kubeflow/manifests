#!/usr/bin/env bash

source hack/utils.sh

if [[ $(basename $PWD) != "manifests" ]]; then
  echo "must be at manifests root directory to run $0"
  exit 1
fi
rm -f tests/*
for i in $(find * -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \; | egrep -v 'tests|hack'); do
  absdir=$(pwd)/$i
  testname=$(get-target-name $absdir)_test.go
  echo generating $testname from $absdir
  ./hack/gen-test-target.sh $absdir > tests/$testname
done
if [[ -d $GOPATH/src/github.com/kubeflow/kubeflow/bootstrap/v2/pkg/kfapp/kustomize ]]; then
  for i in $(ls tests/*_test.go|grep -v kusttestharness_test.go);do
    test=$(basename $i)
    rm $GOPATH/src/github.com/kubeflow/kubeflow/bootstrap/v2/pkg/kfapp/kustomize/$test
  done
  mv tests/* $GOPATH/src/github.com/kubeflow/kubeflow/bootstrap/v2/pkg/kfapp/kustomize
fi
