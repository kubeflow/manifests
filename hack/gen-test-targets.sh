#!/usr/bin/env bash
#
# gen-test-targets will generate units tests under tests/ for all directories that
# have a kustomization.yaml. This script first finds all directories and then calls
# gen-test-target to generate each golang unit test.
# The script is based on kusttestharness_test.go from kubernetes-sigs/pkg/kusttest/kusttestharness.go
#
add_app=''
dry_run=false
exclude_dirs='kfdef|gatekeeper|gcp/deployment_manager_configs|aws/infra_configs'

if [[ $(basename $PWD) != "manifests" ]]; then
  echo "must be at manifests root directory to run $0"
  exit 1
fi

if [[ ! -f hack/utils.sh ]]; then
  echo "$PWD/hack/utils.sh doesn't exist"
  exit 1
fi

source hack/utils.sh

usage () 
{
  echo -e "Usage: $0 [OPTIONS] <directory>\n"\
  'OPTIONS:\n'\
  '  -h | --help       \n'\
  '     | --add-app <name=version>\n'\
  '     | --dry-run\n'
}

findcommand()
{
#  find * -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \; | egrep -v 'doc|tests|hack|plugins'
  _findcommand()
  {
    local branch i
    case "$all" in 
      "false")
        branch=$(lsbranchname)
        for i in $(git diff --name-only origin/${branch}..upstream/master); do
          if [[ -f $i && ! $i =~ ^(doc|tests|hack|plugins) ]]; then   
            echo $(dirname $i)
          fi
        done
        ;;
      "true")
        for i in $(find * -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \;); do
          if [[ $i =~ (base|overlays.*)$ && ! $i =~ ^(doc|tests|hack|plugins) ]]; then
            echo $i
          fi
        done
        ;;
      *)
        echo "Unknown arguments: $@"
        exit 1
        ;;
    esac
  }
  _findcommand $(git branch|grep '^*'|awk '{print $2}') | sort | uniq
}

addapp()
{
  local app=${1%=*} version=${1#*=} appdir=${2}/overlays/application 
#echo 'appdir='$appdir' app='$app' version='$version
  if [[ ! -d $appdir ]]; then
    mkdir -p $appdir
  fi
  cat << KUSTOMIZATION > ${appdir}/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: $app
  app.kubernetes.io/instance: ${app}-${version}
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: $app
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: $version
KUSTOMIZATION

  cat << APPLICATION > ${appdir}/application.yaml
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $app
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: $app
      app.kubernetes.io/instance: ${app}-${version}
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: $app
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: $version
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: Deployment
  descriptor:
    type: $app
    version: v1beta1
    description: ""
    maintainers: []
    owners: []
    keywords:
     - $app
     - kubeflow
    links:
    - description: About
      url: ""
  addOwnerRef: true
APPLICATION
  git add $appdir
}

generate()
{
  local rootdir=$PWD absdir i
  absdir=${rootdir}/$1
#echo 'rootdir='$rootdir' absdir='$absdir 2>&1
  if [[ -n $add_app ]]; then
    absdir=${rootdir}/$1/overlays/application
  fi
  for i in $(find $absdir -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \;); do
#echo 'generate i='$i 2>&1
    if [[ ! $1 =~ ^($exclude_dirs) ]]; then
      testname=$(get-target-name ${i})_test.go
      echo generating $testname from manifests/${i#*manifests/}
      $dry_run || ./hack/gen-test-target.sh $i 1> tests/$testname
      if [[ -n $add_app ]]; then
        $dry_run || git add tests/$testname
      fi
    fi
  done
}

generate-all()
{
  for i in $(findcommand); do
#echo 'generate-all i='$i 2>&1
    generate $i
  done
}

while :
do
  case "$1" in
    -h | --help)
      usage
      exit 0
      ;;
    --add-app)
      shift
      add_app=$1
      shift
      ;;
    --dry-run)
      dry_run=true
      shift
      ;;
    -*)
      echo "Error: Unknown option: $1" >&2
      exit 1
      ;;
    *) 
      break
      ;;
  esac
done

case $# in 
  0)
     generate-all
     ;;
  1)
     if (( $# == 1 )); then
       dir=$1
       if [[ -n $add_app ]]; then
         addapp $add_app $dir
         dir=${dir}/overlays/application
       fi
       generate $dir
     fi
     ;;
  *)
     echo "unknown arguments $@"
     usage
     exit 1
     ;;
esac
