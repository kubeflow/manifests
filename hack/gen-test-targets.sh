#!/usr/bin/env bash
#
# gen-test-targets will generate units tests under tests/ for all directories that
# have a kustomization.yaml. This script first finds all directories and then calls
# gen-test-target to generate each golang unit test.
# The script is based on kusttestharness_test.go from kubernetes-sigs/pkg/kusttest/kusttestharness.go
#
gen_app=''
all=false
changed=false
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

usage() 
{
  echo -e "Usage: $0 [OPTIONS] [<directory>]\n"\
  'OPTIONS:\n'\
  '  -h | --help       \n'\
  '     | --gen-app <name=version> <directory>\n'\
  '  -a | --all\n'\
  '  -c | --changed-only\n'\
  '  -d | --dry-run'
}

usage-extended()
{
  echo ''
  echo 'Examples:'
  echo '1. Generate all unit tests.'
  echo '   $ '$0' --all'
  echo '   generating bootstrap-overlays-application_test.go from manifests/admission-webhook/bootstrap/overlays/application'
  echo '   generating bootstrap-base_test.go from manifests/admission-webhook/bootstrap/base'
  echo '   generating webhook-overlays-application_test.go from manifests/admission-webhook/webhook/overlays/application'
  echo '   generating webhook-base_test.go from manifests/admission-webhook/webhook/base'
  echo '   ...'
  echo '2. Generate an application overlay for spartakus.'
  echo '   $ '$0' --gen-app spartakus=v0.7.0 common/spartakus'
  echo '   mkdir -p common/spartakus/overlays/application'
  echo '   git add common/spartakus/overlays/application'
  echo '   generate common/spartakus/overlays/application'
  echo "3. Update an application overlay's version for admissionwebhook/webhook."
  echo '   $ '$0' --gen-app webhook=v0.7.0 admissionwebhook/webhook'
  echo '   editfile admission-webhook/webhook/overlays/application/application.yaml webhook v0.7.0'
  echo '   editfile admission-webhook/webhook/overlays/application/kustomization.yaml webhook v0.7.0'
  echo '   generate admission-webhook/webhook'
  echo '4. Generate unit tests for just the changed resources on the current branch.'
  echo '   $ '$0' --changed-only'
  echo '   generating webhook-overlays-application_test.go from manifests/admission-webhook/webhook/overlays/application'
  echo '5. Show what unit tests would be generated for just the changed resources on the current branch.'
  echo '   $ '$0' --changed-only --dry-run'
  echo '   generating webhook-overlays-application_test.go from manifests/admission-webhook/webhook/overlays/application'
}

lsbranchname()
{
  git branch|grep '^*'|awk '{print $2}'
}

findremotebranch()
{
  git ls-remote --heads $(git config remote.origin.url) $(lsbranchname) | wc -l
}

findcommand()
{
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
  ! $all && (( $(findremotebranch) == 0 )) && \
  echo "Branch: origin/$(lsbranchname) doesn't exist, push local changes first" >&2 && exit 1
  _findcommand | sort | uniq
}

editfile()
{
  local file=$1 app=$2 version=$3
  if [[ -f $file ]]; then
    ed -s $file <<EDIT_APPLICATION
%sxapp.kubernetes.io/instance: .*xapp.kubernetes.io/instance: ${app}-${version}x
w
%sxapp.kubernetes.io/version: .*xapp.kubernetes.io/version: ${version}x
w
q
EDIT_APPLICATION
  else
    echo "$file doesn't exist"
  fi
}

genapp()
{
  local app=${1%=*} version=${1#*=} appdir=${2}/overlays/application cmd
#echo 'appdir='$appdir' app='$app' version='$version
  if [[ ! -d $appdir ]]; then
    cmd="mkdir -p $appdir"
    $dry_run && cmd='echo '$cmd
    eval $cmd
    $dry_run || cat << KUSTOMIZATION > ${appdir}/kustomization.yaml
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

    $dry_run || cat << APPLICATION > ${appdir}/application.yaml
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

    cmd="git add $appdir"
    $dry_run && cmd='echo '$cmd
    eval $cmd
  else
    for i in $appdir/{application.yaml,kustomization.yaml}; do 
      cmd="editfile $i ${app} ${version}"
      $dry_run && cmd='echo '$cmd
      eval $cmd
    done
  fi
}

generate()
{
  local rootdir=$PWD absdir i
  absdir=${rootdir}/$1
#echo 'rootdir='$rootdir' absdir='$absdir
  if [[ -n $gen_app ]]; then
    absdir=${rootdir}/$1/overlays/application
  fi
  for i in $(find $absdir -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \;); do
#echo 'generate i='$i 2>&1
    if [[ ! $1 =~ ^($exclude_dirs) ]]; then
      testname=$(get-target-name ${i})_test.go
      echo generating $testname from manifests/${i#*manifests/}
      $dry_run || ./hack/gen-test-target.sh $i 1> tests/$testname
      if [[ -n $gen_app ]]; then
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
    -h)
      usage
      exit 0
      ;;
    --help)
      usage && usage-extended
      exit 0
      ;;
    --gen-app)
      shift
      gen_app=$1
      shift
      ;;
    -a | --all)
      all=true
      shift
      ;;
    -c | --changed-only)
      changed=true
      shift
      ;;
    -d | --dry-run)
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
     ( $all || $changed ) && generate-all || \
     ( echo -e 'Either --all or --changed-only or directory required\n' && usage && exit 1 )
     ;;
  1)
     if (( $# == 1 )); then
       if [[ -n $gen_app ]]; then
         genapp $gen_app $1
       fi
       cmd="generate $1"
       $dry_run && cmd='echo '$cmd
       eval $cmd
     fi
     ;;
  *)
     echo "Unknown arguments: $@"
     exit 1
     ;;
esac
