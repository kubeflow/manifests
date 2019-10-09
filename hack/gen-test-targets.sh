#!/usr/bin/env bash
#
# gen-test-targets will generate units tests under tests/ for all directories that
# have a kustomization.yaml. This script first finds all directories and then calls
# gen-test-target to generate each golang unit test.
# The script is based on kusttestharness_test.go from kubernetes-sigs/pkg/kusttest/kusttestharness.go
#
add_app=''
source hack/utils.sh
usage () 
{
  echo -e "Usage: $0 [OPTIONS] <directory>\n"\
  'OPTIONS:\n'\
  '  -h | --help       \n'\
  '     | --add-app <name=version>\n'
}

findcommand()
{
#  find * -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \; | egrep -v 'doc|tests|hack|plugins'
  _findcommand()
  {
    local branch=$1
    for i in $(git diff --name-only origin/${branch}..upstream/master | egrep -v 'doc|tests|hack|plugins');do 
      if [[ -f $i ]]; then   
        echo $(dirname $i); 
      fi
    done
  }
  _findcommand $(git branch|grep '^*'|awk '{print $2}') | sort | uniq
}

addapp()
{
  local app=${1%=*} version=${1#*=} appdir=${2}/overlays/application 
echo 'appdir='$appdir' app='$app' version='$version
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
      url: link
  addOwnerRef: true
APPLICATION
}

generate()
{
  local rootdir=$(pwd) absdir i
  absdir=${rootdir}/$1
echo 'rootdir='$rootdir' absdir='$absdir
  for i in $(find $absdir -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \;); do
    if [[ ! $i =~ overlays/test$ ]]; then
      testname=$(get-target-name ${i})_test.go
      echo generating $testname from manifests/${i#*manifests/}
      ./hack/gen-test-target.sh $i 1> tests/$testname
    fi
  done
}

generate-all()
{
  if [[ $(basename $PWD) != "manifests" ]]; then
    echo "must be at manifests root directory to run $0"
    exit 1
  fi
  EXCLUDE_DIRS=( "kfdef" "gatekeeper" "gcp/deployment_manager_configs" "aws/infra_configs" )
  source hack/utils.sh
  for i in $(findcommand); do
    exclude=false
    for item in "${EXCLUDE_DIRS[@]}"
    do
      #https://stackoverflow.com/questions/2172352/in-bash-how-can-i-check-if-a-string-begins-with-some-value
      # Check if item is a prefix of i
      if [[ "$i" == "$item"* ]]; then
        exclude=true
      fi
    done
    if $exclude; then
      continue
    fi
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
       if [[ -n $add_app ]]; then
         addapp $add_app $1
       fi
       generate $1
     fi
     ;;
  *)
     echo "unknown arguments $@"
     usage
     exit 1
     ;;
esac
