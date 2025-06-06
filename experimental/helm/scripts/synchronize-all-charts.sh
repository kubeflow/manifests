#!/usr/bin/env bash
# Master script to sync all upstream charts for Kubeflow AIO Helm chart

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HELM_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$HELM_DIR/kubeflow"

COMPONENTS=(
    "spark-operator"
    # Add more components as we implement them
    # "training-operator"
    # "cert-manager"
    # "istio"
    # "oauth2-proxy"
    # "dex"
)

for component in "${COMPONENTS[@]}"; do
    sync_script="$SCRIPT_DIR/sync-${component}.sh"
    [ -f "$sync_script" ] && bash "$sync_script"
done

cd "$CHART_DIR"
helm template kubeflow . --debug --dry-run > /dev/null 