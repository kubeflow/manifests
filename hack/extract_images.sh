#!/usr/bin/env bash

#| Working Group         	| Directories                                                                                                                                                                                                                                                                                                                      	|
#|-----------------------	|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------	|
#| AutoML(Katib)         	| ../apps/katib                                                                                                                                                                                                                                                                                                                       	|
#| Pipelines             	| ../apps/pipeline/upstream                                                                                                                                                                                                                                                                                                           	|
#| Training              	| ../apps/training-operator/upstream                                                                                                                                                                                                                                                                                                 	|
#| Manifests             	| ../common                                                                                                                                                                                                                                                                                                                          	|
#| Notebooks/Workbenches 	| ../apps/admission-webhook/upstream<br>../apps/centraldashboard/upstream<br>../apps/jupyter/jupyter-web-app/upstream<br>../apps/volumes-web-app/upstream<br>../apps/tensorboard/tensorboards-web-app/upstream<br>../apps/profiles/upstream<br>../apps/jupyter/notebook-controller/upstream<br>../apps/tensorboard/tensorboard-controller/upstream 	|

# The script reports:
# 1. Images used by the Kubeflow Working Groups
# 2. All images used by Kubeflow
# The reported image lists are saved in respective files under ../doc directory

# Future release process enhancements may include an automatic image inventory scan.
# The reported image list can also be used for image vulnerability scanning and managing licenses

version=latest
images=()

declare -A wg_dirs=(
  [automl]="../apps/katib"
  [pipelines]="../apps/pipeline/upstream"
  [training]="../apps/training-operator/upstream"
  [manifests]="../common"
  [notebooks]="../apps/admission-webhook/upstream ../apps/centraldashboard/upstream ../apps/jupyter/jupyter-web-app/upstream ../apps/volumes-web-app/upstream ../apps/tensorboard/tensorboards-web-app/upstream ../apps/profiles/upstream ../apps/jupyter/notebook-controller/upstream ../apps/tensorboard/tensorboard-controller/upstream"
)

declare -A wg_exclude_dirs=(
#  [automl]="*/manager/* */default/* */crd/* */rbac/* */components/*"
  [notebooks]="*/manager/* */default/* */crd/* */rbac/* */components/*"
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
  output_file="../docs/kf${version}_${wg}_images.txt"
  printf "%s\n" "${images[@]}" > "$output_file"
  echo "File ${output_file} successfully created"
}

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
