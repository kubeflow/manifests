#!/usr/bin/env bash

# The script extracts all images from specific manifests sub-directories.
# Future release process enhancements may include an automatic image inventory scan.
# The reported image list can later be used for image vulnerability scanning and managing license risks management

#VERSION=1.7.0
output_file="../docs/kf${VERSION}_images.txt"
declare -a dirs=("../apps" "../common" "../example" "../contrib/metacontroller"
                 "../contrib/seldon" "../contrib/bentoml" )
rm -f .tmp
# Iterate over all files with names: 'kustomization.yaml', 'kustomization.yml', 'Kustomization' found recursively in the provided list of directories
for F in $(find "${dirs[@]}" \( -name kustomization.yaml   -o -name kustomization.yml -o -name Kustomization \)); do

  dir=$(dirname -- "$F")
  # Generate k8s resources specified in 'dir' using the 'kustomize build' command.
  # Log the 'dir' name where the 'kustomize build' command fails.
  kbuild=$(kustomize build "$dir")
  return_code=$?
  if [ $return_code -ne 0 ]; then
    printf 'ERROR:\t Failed \"kustomize build\" command for directory: %s. See error above\n' "$dir"
     continue
  fi
    # Grep the output of 'kustomize build' command for 'image:' and '- image' lines and return just the image itself
    # Redirect the output to '.tmp' file
  grep '\-\?\s\image:'<<<"$kbuild" | sed -re 's/\s-?\simage: *//;s/^[ \t]*//g' | sed '/^$/d;/{/d' >> .tmp
done

sort .tmp | uniq > "$output_file"
rm -f .tmp

echo "File ${output_file} successfully created"
