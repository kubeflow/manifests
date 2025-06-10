#!/usr/bin/env bash
# Master script to sync all upstream charts for Kubeflow AIO Helm chart

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HELM_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$HELM_DIR/kubeflow"

COMPONENTS=(
    "spark-operator"
    "cert-manager"
    # Add more components as we implement them
    # "training-operator"
    # "istio"
    # "oauth2-proxy"
    # "dex"
)

for component in "${COMPONENTS[@]}"; do
    sync_script="$SCRIPT_DIR/synchronize-${component}.sh"
    if [ -f "$sync_script" ]; then
        echo "Syncing $component..."
        bash "$sync_script"
    fi
done

cd "$CHART_DIR"
helm template kubeflow . --debug --dry-run > /dev/null && echo "Success: AIO chart templates correctly!" 