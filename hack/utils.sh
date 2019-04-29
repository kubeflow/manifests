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
