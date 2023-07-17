#!/usr/bin/env bash

# The script reports:
# 1. Images used by the Kubeflow Working Groups
# 2. All images used by Kubeflow
# The reported image lists are saved in respective files under ../docs/image_lists directory
# The script must be executed from the `hack` folder as it use relative paths

# Future release process enhancements may include an automatic image inventory scan, generating SBOM files
# vulnerability scanning and managing licenses.

images=()

declare -A wg_dirs=(
  [automl]="../apps/katib/upstream/installs"
  [pipelines]="../apps/pipeline/upstream/env ../apps/kfp-tekton/upstream/env"
  [training]="../apps/training-operator/upstream/overlays"
  [manifests]="../common ../example"
  [workbenches]="../apps/admission-webhook/upstream/overlays ../apps/centraldashboard/upstream/overlays ../apps/jupyter/jupyter-web-app/upstream/overlays ../apps/volumes-web-app/upstream/overlays ../apps/tensorboard/tensorboards-web-app/upstream/overlays ../apps/profiles/upstream/overlays ../apps/jupyter/notebook-controller/upstream/overlays ../apps/tensorboard/tensorboard-controller/upstream/overlays"
  [serving]="../contrib/kserve"
)

declare -A wg_exclude_dirs=(
  [workbenches]="*/manager/* */default/* */crd/* */rbac/* */components/*"
)

get_wg_ignored_dirs() {
  local wg=$1
  # Check if the key exists in the map
  if [[ ${wg_exclude_dirs[$wg]+_} ]]; then
    # Split the string into an array using space as the delimiter
    IFS=" " read -ra values <<< "${wg_exclude_dirs[$wg]}"
    echo "${values[@]}"
  else
    echo ""
  fi
}

# Build the 'find' command dynamically
# example: find ../apps/katib ( -name kustomization.yaml   -o -name kustomization.yml -o -name Kustomization \) ! -path '*./xxx/*' ! -path '*/components/*'
construct_find_command(){
  find_command="find $1 \( -name kustomization.yaml   -o -name kustomization.yml -o -name Kustomization \) "
  for ignore_dir in $2; do
    find_command+="! -path \"$ignore_dir\" "
  done
  echo "$find_command"
}

save_images() {
  wg=${1:-""}
  shift
  local images=("$@")
  output_file="../docs/image_lists/kf_${version}_${wg}_images.txt"
  printf "%s\n" "${images[@]}" > "$output_file"
  echo "File ${output_file} successfully created"
}

validate_semantic_version() {
  local version="${1:-"latest"}"

  local regex="^[0-9]+\.[0-9]+\.[0-9]+$"  # Regular expression for semantic version pattern
  if [[ $version  =~ $regex || $version = "latest" ]]; then
      echo "$version"
  else
      echo "Invalid semantic version: '$version'"
      return 1
  fi
}

if ! version=$(validate_semantic_version "$1") ; then
    echo "$version. Exiting script."
    exit 1
fi

echo "Running the script using Kubeflow version: $version"

for wg in "${!wg_dirs[@]}"; do
  ignored_dirs=$(get_wg_ignored_dirs "$wg")
  declare -a dirs=(${wg_dirs[$wg]})
  wg_images=()
  for (( i=0; i<"${#dirs[@]}"; i++ )); do
    find_command=$(construct_find_command "${dirs[$i]}" "$ignored_dirs")
    kustomization_files=($(eval "$find_command"))
    for F in "${kustomization_files[@]}"; do
        dir=$(dirname -- "$F")
        # Generate k8s resources specified in 'dir' using the 'kustomize build' command.
        kbuild=$(kustomize build "$dir")
        return_code=$?
        if [ $return_code -ne 0 ]; then
          printf 'ERROR:\t Failed \"kustomize build\" command for directory: %s. See error above\n' "$dir"
          continue
        fi
        # Grep the output of 'kustomize build' command for 'image:' and '- image' lines and return just the image itself
        mapfile kimages -t  <<< "$(grep '\-\?\s\image:'<<<"$kbuild" | sed -re 's/\s-?\simage: *//;s/^[ \t]*//g' | sed '/^$/d;/{/d' )"
        wg_images+=("${kimages[@]}")
    done
  done
  uniq_wg_images=($(echo "${wg_images[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
  images+=(${uniq_wg_images[@]})
  save_images "${wg}" "${uniq_wg_images[@]}"
done

uniq_images=($(echo "${images[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
save_images "all" "${uniq_images[@]}"
