#
# utils.sh has common scripts used the gen-test-targets, get-test-target and gen-tree.
# 

#
# manifests-tree will produce a listing that can be included in the README.md that shows 
# what directories hold kustomization.yamls. 
#
tmpfile=""
cleanup()
{
  if [[ -f $tmpfile ]]; then
    rm -f $tmpfile
  fi
}
trap cleanup EXIT

manifests-tree()
{
   local dir='*'
   if (( $# >= 1 )); then
     dir=$1
     shift
   fi
   tmpfile=$(mktemp -q -t tree)
   for i in $(find * -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \; | egrep -v 'docs|tests|hack'); do
     d=$(dirname $i) 
     b=$(basename $i)
     echo /manifests/$d/🎯$b >> $tmpfile
   done
   cat $tmpfile | tree $@ -N --fromfile --noreport
}

#
# get-target will return the 'root' of the manifest given the full path to where the kustomization.yaml is.
# For example
#
# tf-job-operator
# ├── base
# └── overlays
#     ├── cluster
#     ├── cluster-gangscheduled
#     ├── namespaced
#     └── namespaced-gangscheduled
#
# Given the path /manifests/tf-training/tf-job-operator/overlays/namespaced-gangscheduled
# gen-target will return /manifests/tf-training/tf-job-operator
#
# Given the path /manifests/tf-training/tf-job-operator/base
# gen-target will return /manifests/tf-training/tf-job-operator
#
get-target()
{
  local b=$(basename $1)
  case $b in
    base)
      echo $(dirname $1)
      ;;
    *)
      echo $(dirname $(dirname $1))
      ;;
  esac
}

#
# get-target will return the basename of the manifest given the full path to where the kustomization.yaml is.
# For example
#
# tf-job-operator
# ├── base
# └── overlays
#     ├── cluster
#     ├── cluster-gangscheduled
#     ├── namespaced
#     └── namespaced-gangscheduled
#
# Given the path /manifests/tf-training/tf-job-operator/overlays/namespaced-gangscheduled
# gen-target will return tf-job-operator
#
# Given the path /manifests/tf-training/tf-job-operator/base
# gen-target will return tf-job-operator
#
get-target-name()
{
  local b=$(basename $1)
  case $b in
    base)
      echo $(basename $(dirname $1))
      ;;
    *)
      echo $(basename $(dirname $(dirname $1)))
      ;;
  esac
}

#
# get-target-dirname will return the dirs between the root and the kustomization.yaml
# For example
#
# tf-job-operator
# ├── base
# └── overlays
#     ├── cluster
#     ├── cluster-gangscheduled
#     ├── namespaced
#     └── namespaced-gangscheduled
#
# Given the path /manifests/tf-training/tf-job-operator/overlays/namespaced-gangscheduled
# gen-target-dirname will return overlays/namespaced-gangscheduled
#
# Given the path /manifests/tf-training/tf-job-operator/base
# gen-target will return base
#
get-target-dirname()
{
  local b=$(basename $1)
  case $b in
    base)
      echo base
      ;;
    *)
      echo overlays/$b
      ;;
  esac
}
