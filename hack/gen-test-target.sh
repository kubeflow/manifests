#!/usr/bin/env bash

kebab-case-2-PascalCase()
{
  local a=$1 b='' array
  IFS='-' read -r -a array <<< "$a"
  for element in "${array[@]}"
  do
    part="${element^}"
    b+=$part
  done
  echo $b
}

gen-target-start()
{
  local dir=$(dirname $1) target fname
  fname=/manifests${dir##*/manifests}
  target=$(kebab-case-2-PascalCase $(basename $dir))

  echo 'package kustomize_test'
  echo ''
  echo 'import ('
  echo '  "testing"'
  echo ')'
  echo ''
  echo 'func write'$target'(th *KustTestHarness) {'
}

gen-target-end()
{
  echo '}'
}

gen-target()
{
  local directory=$1
  gen-target-start $1
  for i in $(ls $directory); do
    file=$(basename $i)
    case $file in
      "kustomization.yaml")
        gen-target-kustomization $file $directory
        ;;
      *.yaml)
        gen-target-resource $file $directory
        ;;
      params.env)
        gen-target-resource $file $directory
        ;;
      *)
        ;;
    esac
  done
  gen-target-end
}

gen-target-kustomization()
{
  local file=$1 dir=$2 fname kname
  fname=/manifests${dir##*/manifests}
  kname=${fname%/kustomization.yaml}
  echo '  th.writeK("'$kname'", `'
  cat $dir/$file
  echo '`)'

}

gen-target-resource()
{
  local file=$1 dir=$2 fname
  fname=/manifests${dir##*/manifests}/$file

  echo '  th.writeF("'$fname'", `'
  cat $dir/$file 
  echo '`)'
}

gen-expected-start()
{
  echo  '  th.assertActualEqualsExpected(m, `'
}

gen-expected-end()
{
  echo  '`)'
}

gen-expected()
{
  gen-expected-start
  cd $1
  kustomize build
  cd - >/dev/null
  gen-expected-end
}

gen-test-case()
{
  local base=$(basename $1) dir=$(dirname $1) target fname
  fname=/manifests${dir##*/manifests}/$base
  target=$(kebab-case-2-PascalCase $(basename $dir))

  gen-target $1
  echo ''
  echo 'func Test'$target'(t *testing.T) {'
  echo '  th := NewKustTestHarness(t, "'$fname'")'
  echo '  write'$target'(th)'
  echo '  m, err := th.makeKustTarget().MakeCustomizedResMap()'
  echo '  if err != nil {'
  echo '    t.Fatalf("Err: %v", err)'
  echo '  }'
  gen-expected $1
  echo '}'
}

gen-test-case $1
