#!/usr/bin/env bash
# Script to test Helm charts

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
HELM_DIR=$(dirname "$SCRIPT_DIR")

echo "Testing Helm charts in ${HELM_DIR}"

test_chart() {
    local chart_path=$1
    local chart_name=$(basename "$chart_path")
    
    echo "Testing chart: $chart_name"
    helm lint "$chart_path"
    
    if [ -f "$chart_path/Chart.yaml" ] && grep -q "dependencies:" "$chart_path/Chart.yaml"; then
        helm dependency update "$chart_path"
    fi
    
    helm template "test-$chart_name" "$chart_path" > /dev/null
    echo "✓ $chart_name"
}

for chart in "$HELM_DIR"/charts/*/; do
    if [ -f "$chart/Chart.yaml" ]; then
        test_chart "$chart"
    fi
done

if [ -f "$HELM_DIR/kubeflow/Chart.yaml" ]; then
    test_chart "$HELM_DIR/kubeflow"
fi

echo "All charts tested successfully" 