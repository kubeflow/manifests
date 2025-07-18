#!/usr/bin/env bash
# Compare Helm vs Kustomize manifests for all KServe Models Web App scenarios

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

SCENARIOS=(
    "base"
    "kubeflow"
)

echo "KServe Models Web App Helm vs Kustomize Comparison"
echo "Testing scenarios: ${SCENARIOS[*]}"

declare -a PASSED_SCENARIOS=()
declare -a FAILED_SCENARIOS=()

for scenario in "${SCENARIOS[@]}"; do
    echo "Testing scenario: $scenario"
    
    if "$SCRIPT_DIR/models_web_app_compare.sh" "$scenario"; then
        echo "PASSED: $scenario"
        PASSED_SCENARIOS+=("$scenario")
    else
        echo "FAILED: $scenario"
        FAILED_SCENARIOS+=("$scenario")
    fi
done

echo "FINAL RESULTS"
echo "Passed scenarios (${#PASSED_SCENARIOS[@]}/${#SCENARIOS[@]}):"
for scenario in "${PASSED_SCENARIOS[@]}"; do
    echo "  - $scenario"
done

if [ ${#FAILED_SCENARIOS[@]} -gt 0 ]; then
    echo "Failed scenarios (${#FAILED_SCENARIOS[@]}/${#SCENARIOS[@]}):"
    for scenario in "${FAILED_SCENARIOS[@]}"; do
        echo "  - $scenario"
    done
    exit 1
else
    exit 0
fi 