manifests-tree()
{
   local dir='*'
   if (( $# >= 1 )); then
     dir=$1
     shift
   fi
   find $dir -type d  | egrep -v 'tests|hack' | tree $@ --fromfile | sed 's/─ base/🎯base/g' | sed 's/─ gcp/🎯gcp/g' | sed 's/─ \(cluster*\)/─ 🎯\1/g' | sed 's/─ \(namespaced*\)/─ 🎯\1/g'
}

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
