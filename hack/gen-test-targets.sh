#!/usr/bin/env bash

if [[ $(basename $PWD) != "manifests" ]]; then
  echo "must be at manifests root directory to run $0"
  exit 1
fi

source hack/utils.sh
rm -f $(ls tests/*_test.go | grep -v kusttestharness_test.go)
for i in $(find * -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \; | egrep -v 'tests|hack'); do
  absdir=$(pwd)/$i
  testname=$(get-target-name $absdir)_test.go
  echo generating $testname from $absdir
  ./hack/gen-test-target.sh $absdir > tests/$testname
done
