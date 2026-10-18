#!/bin/bash
set -euo pipefail

profile_namespace=${1:-kubeflow-user-example-com}
registry_endpoint="http://model-registry-service.${profile_namespace}.svc.cluster.local:8080/api/model_registry/v1alpha3/registered_models"
probe_suffix="${RANDOM}"
ui_probe="registry-ui-egress-${probe_suffix}"
control_probe="registry-egress-control-${probe_suffix}"

cleanup() {
    kubectl -n kubeflow delete pod "$ui_probe" "$control_probe" --ignore-not-found --wait=false
}
trap cleanup EXIT

for probe_name in "$ui_probe" "$control_probe"; do
    application_label=registry-egress-control
    if [[ "$probe_name" == "$ui_probe" ]]; then
        application_label=model-registry-ui
    fi
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${probe_name}
  namespace: kubeflow
  labels:
    app: ${application_label}
spec:
  restartPolicy: Never
  automountServiceAccountToken: false
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: registry-egress-probe
    image: busybox:1.37.0
    command: ["sleep", "600"]
    resources:
      requests:
        cpu: 10m
        memory: 16Mi
      limits:
        memory: 64Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
EOF
done
kubectl -n kubeflow wait --for=condition=Ready "pod/$ui_probe" "pod/$control_probe" --timeout=180s
profile_token=$(kubectl -n "$profile_namespace" create token default-editor --duration=10m)

assert_response() {
    local probe_name=$1
    local expected_status=$2
    shift 2
    local response
    response=$(kubectl -n kubeflow exec "$probe_name" -c registry-egress-probe -- \
        wget -S -O /dev/null -T 20 "$@" "$registry_endpoint" 2>&1) || true
    if ! grep -Eq "HTTP/1[.][01] ${expected_status} " <<< "$response"; then
        printf 'FAIL: %s expected HTTP %s\n%s\n' "$probe_name" "$expected_status" "$response" >&2
        return 1
    fi
    printf 'PASS: %s returned HTTP %s\n' "$probe_name" "$expected_status"
}

assert_response "$ui_probe" 200 --header="Authorization: Bearer $profile_token"
assert_response "$ui_probe" 403
assert_response "$control_probe" 403 --header="Authorization: Bearer $profile_token"
printf 'Registry cross-namespace egress and authorization checks passed.\n'
