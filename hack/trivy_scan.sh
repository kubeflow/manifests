# !/usr/bin/env bash

# The script:
# 1. Extract all the images used by the Kubeflow Working Groups
# - The reported image lists are saved in respective files under ../docs/image_lists directory
# 2. Scan the reported images using Trivy for security vulnerabilities
# - Scanned reports will be saved in JSON format inside ../image_lists/security_scan_reports/ folder for each Working Group
# 3. The script will also generate a summary of the security scan reports with severity counts for each Working Group with images
# - Summary of security counts with images a JSON file inside ../image_lists/summary_of_severity_counts_for_WG folder
# 4. Generate a summary of the security scan reports
# - The summary will be saved in JSON format inside ../image_lists/summary_of_severity_counts_for_WG folder
# 5. Before run this file you have to 
#    1. Install kustomize 
#       - sudo apt install snapd
#       - sudo snap install kustomize
#    2. Install trivy
#       - sudo apt install snapd
#       - sudo snap install trivy
#    3. Install jq
#       - sudo apt install jq
#    4. Install Python
#    5. Install prettytable
#       - pip install prettytable

# The script must be executed from the hack folder as it use relative paths

echo "Extracting Images"
images=()

declare -A wg_dirs=(
  [automl]="../apps/katib/upstream/installs"
  [pipelines]="../apps/pipeline/upstream/env ../apps/kfp-tekton/upstream/env"
  [training]="../apps/training-operator/upstream/overlays"
  [manifests]="../common/cert-manager/cert-manager/base ../common/cert-manager/kubeflow-issuer/base ../common/istio-1-17/istio-crds/base ../common/istio-1-17/istio-namespace/base ../common/istio-1-17/istio-install/overlays/oauth2-proxy ../common/oidc-client/oauth2-proxy/overlays/m2m-self-signed ../common/dex/overlays/oauth2-proxy ../common/knative/knative-serving/overlays/gateways ../common/knative/knative-eventing/base ../common/istio-1-17/cluster-local-gateway/base ../common/kubeflow-namespace/base ../common/kubeflow-roles/base ../common/istio-1-17/kubeflow-istio-resources/base"
  [workbenches]="../apps/pvcviewer-controller/upstream/base ../apps/admission-webhook/upstream/overlays ../apps/centraldashboard/upstream/overlays/oauth2-proxy ../apps/jupyter/jupyter-web-app/upstream/overlays ../apps/volumes-web-app/upstream/overlays ../apps/tensorboard/tensorboards-web-app/upstream/overlays ../apps/profiles/upstream/overlays ../apps/jupyter/notebook-controller/upstream/overlays ../apps/tensorboard/tensorboard-controller/upstream/overlays"
  [serving]="../contrib/kserve - ../contrib/kserve/models-web-app/overlays/kubeflow"
  [model-registry]="../apps/model-registry/upstream"

)

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
  declare -a dirs=(${wg_dirs[$wg]})
  wg_images=()
  for (( i=0; i<"${#dirs[@]}"; i++ )); do
    for F in $(find "${dirs[$i]}" \( -name kustomization.yaml   -o -name kustomization.yml -o -name Kustomization \)); do
        dir=$(dirname -- "$F")
        # Generate k8s resources specified in 'dir' using the 'kustomize build' command.
        kbuild=$(kustomize build "$dir")
        return_code=$?
        if [ $return_code -ne 0 ]; then
          printf 'ERROR:\t Failed \"kustomize build\" command for directory: %s. See error above\n' "$dir"
          continue
        fi
        # Grep the output of 'kustomize build' command for 'image:' and '- image' lines and return just the image itself
        mapfile kimages -t  <<< "$(grep '\-\?\s\image:'<<<"$kbuild" | sed -re 's/\s-?\simage: *//;s/^[ \t]*//g' | sed '/^\$/d;/{/d' )"
        wg_images+=("${kimages[@]}")
    done
  done
  uniq_wg_images=($(echo "${wg_images[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
  images+=(${uniq_wg_images[@]})
  save_images "${wg}" "${uniq_wg_images[@]}"
done

uniq_images=($(echo "${images[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
save_images "all" "${uniq_images[@]}"


# Directory containing the text files
DIRECTORY="../docs/image_lists"

# Directory to save security scan reports
SCAN_REPORTS_DIR="${DIRECTORY}/security_scan_reports"
mkdir -p "$SCAN_REPORTS_DIR"

#Directory to save severity counts with images for the WG
ALL_SEVERITY_COUNTS="${DIRECTORY}/severity_counts_with_images_for_WG"
mkdir -p "$ALL_SEVERITY_COUNTS"

#Directory to save summary of the severity counts of the WG
SUMMARY_OF_SEVERITY_COUNTS="${DIRECTORY}/summary_of_severity_counts_for_WG"
mkdir -p "$SUMMARY_OF_SEVERITY_COUNTS"

echo "Started scanning images"

files=($(find "$DIRECTORY" -type f -name "*.txt" ! -name "kf_latest_all_images.txt"))

# Loop through each text file in the specified directory
for file in "${files[@]}"; do

    echo "Scanning images in $file"

    # Extract the base name of the file (without the directory and extension)
    file_base_name=$(basename "$file" .txt)

    # Directory to save reports for this specific file
    file_reports_dir="${SCAN_REPORTS_DIR}/${file_base_name}"
    mkdir -p "$file_reports_dir"

    # Directory to save securtiy count
    severity_count="${file_reports_dir}/severity_counts"
    mkdir -p "$severity_count"

    while IFS= read -r line; do
        # Extract the image name (removing the tag/version)
        image_name=$(echo "$line" | cut -d':' -f1)
        image_tag=$(echo "$line" | cut -d':' -f2)

        # Set default scan file name
        image_name_scan="${image_name##*/}"

        # Append tag to the scan file name if it exists
        if [ -n "$image_tag" ]; then
                image_name_scan="${image_name_scan}_${image_tag}"
        fi

        echo "Scanning $image_name_scan"

        trivy image --format json --output "${file_reports_dir}/${image_name_scan}_scan.json" "$line"
        if [ $? -ne 0 ]; then 
            echo "Error scanning $image_name:$image_tag"
        else
        # Check if results exist in the scan file (before processing)
          is_json_empty=$(jq -r '.Results // false' "${file_reports_dir}/${image_name_scan}_scan.json")

          if [[ "$is_json_empty" == "false" ]]; then
                  echo "No vulnerabilities found in $image_name:$image_tag"
          else
              # Filter results to include only elements with vulnerabilities
              results=$(jq -r '.Results? | .[] | select(.Vulnerabilities) | .Vulnerabilities | length > 0' "${file_reports_dir}/${image_name_scan}_scan.json") 
              if [[ "$results" == "" || "$results" == "false" ]]; then
                      echo "The vulnerability detection may be insufficient because security updates are not provided for $image_name:$image_tag"
              else
                      # Count the number of vulnerabilities by severity
                      severity_counts=$(jq 'reduce (.Results[].Vulnerabilities? // [])[] as $v ({"LOW": 0, "MEDIUM": 0, "HIGH": 0, "CRITICAL": 0}; .[$v.Severity]+=1)' "${file_reports_dir}/${image_name_scan}_scan.json")
                      report=$(jq -n --arg image "$line" --argjson counts "$severity_counts" '{image: $image, severity_counts: $counts}')
                      echo "$report" > "${severity_count}/${image_name_scan}_severity_report.json"
              fi
          fi
        fi

    done < "$file"


    # Combine all the JSON files into a single file with severity counts for all images
    json_dir="${severity_count}"

    output_file="${ALL_SEVERITY_COUNTS}/${file_base_name}.json"

    if [ -z "$(ls -A ${json_dir})" ]; then
      echo "No JSON files found in '$json_dir'. Skipping combination."
    else
      jq -s '. | { "data": [.[]]}' ${json_dir}/*.json > "$output_file"

      if [[ $? -eq 0 ]]; then
        echo "JSON files successfully combined into '$output_file'"
      else
        echo "Error: Failed to combine JSON files."
      fi
    fi

done

#Directory containing the summary of severity counts related to images
severity_dir="${ALL_SEVERITY_COUNTS}"

summary_file="${SUMMARY_OF_SEVERITY_COUNTS}/severity_summary_in_json_format.json"

# Initialize counters
total_images=0
total_low=0
total_medium=0
total_high=0
total_critical=0

# Initialize a variable to hold the final JSON
merged_data='{}'

  # Loop through each JSON file
for file in "$severity_dir"/*.json; do
  # Get filename without extension
  filename=$(basename "$file" .json)
  filename="${filename##kf_latest_}"
  filename="${filename%%_images}"
  filename="${filename^}"
   # Process the JSON file
  data=$(jq -r '.data[] | {LOW: .severity_counts.LOW, MEDIUM: .severity_counts.MEDIUM, HIGH: .severity_counts.HIGH, CRITICAL: .severity_counts.CRITICAL}' "$file")

  # Check if data is empty
  if [[ -z "$data" ]]; then
    data="{\"LOW\": 0, \"MEDIUM\": 0, \"HIGH\": 0, \"CRITICAL\": 0}"
  fi

  # Extract counts for this file
  image_count=$(jq '.data | length' "$file")
  low=$(jq -r '.LOW' <<< "$data" | awk '{s+=$1} END {print s}')
  medium=$(jq -r '.MEDIUM' <<< "$data" | awk '{s+=$1} END {print s}')
  high=$(jq -r '.HIGH' <<< "$data" | awk '{s+=$1} END {print s}')
  critical=$(jq -r '.CRITICAL' <<< "$data" | awk '{s+=$1} END {print s}')

  # Update the total counts
  total_images=$((total_images + image_count))
  total_low=$((total_low + low))
  total_medium=$((total_medium + medium))
  total_high=$((total_high + high))
  total_critical=$((total_critical + critical))


  # Create the output for this file
  file_data=$(jq -n --arg images "$image_count" --arg low "$low" --arg medium "$medium" --arg high "$high" --arg critical "$critical" '{
    "images": ($images | tonumber),
    "LOW": ($low | tonumber),
    "MEDIUM": ($medium | tonumber),
    "HIGH": ($high | tonumber),
    "CRITICAL": ($critical | tonumber)
  }')


  # Append to all_data JSON string
  all_data="{\"$filename\": $file_data,\"total\": { \"images\": $total_images, \"LOW\": $total_low, \"MEDIUM\": $total_medium, \"HIGH\": $total_high, \"CRITICAL\": $total_critical }}"
  merged_data=$(jq -s '.[0] * .[1]' <(echo "$merged_data") <(echo "$all_data"))

done

# Write the final output to a file
echo "${merged_data}"  > "$summary_file"

echo "Summary written to: $summary_file"

echo "Image scanning completed. Reports are saved in ${SCAN_REPORTS_DIR}"

# Run the Python script
python3 table_generate_for_security_results.py