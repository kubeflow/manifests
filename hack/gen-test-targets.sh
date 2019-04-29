#!/usr/bin/env bash

if [[ $(basename $PWD) != "manifests" ]]; then
  echo "must be at manifests root directory to run $0"
  exit 1
fi
rm -f tests/*
for i in $(find * -name base -a -type d); do
  dir=$(dirname $i)
  dirname=$(basename $dir)
  absdir=$(pwd)/$i
  ./hack/gen-test-target.sh $absdir > tests/${dirname}_test.go
done
if [[ -d $GOPATH/src/github.com/kubeflow/kubeflow/bootstrap/v2/pkg/kfapp/kustomize ]]; then
  for i in $(ls tests/*_test.go|grep -v kusttestharness_test.go);do
    test=$(basename $i)
    rm $GOPATH/src/github.com/kubeflow/kubeflow/bootstrap/v2/pkg/kfapp/kustomize/$test
  done
  mv tests/* $GOPATH/src/github.com/kubeflow/kubeflow/bootstrap/v2/pkg/kfapp/kustomize
fi
