#!/bin/bash
set -euo pipefail

# This fixture deliberately changes routing and mesh policies in an ephemeral
# KinD test cluster. Never run it against a shared installation.
[[ $(kubectl config current-context) == kind-* ]] || {
    echo 'ERROR: the shared UI regression requires an isolated KinD test cluster.' >&2
    exit 1
}
profile_namespace=${1:-kubeflow-user-example-com}
gateway_endpoint=${GATEWAY_ENDPOINT:-http://localhost:8080}
registry_endpoint="http://model-registry-service.${profile_namespace}.svc.cluster.local:8080/api/model_registry/v1alpha3/registered_models"
ui_route="${gateway_endpoint}/model-registry/api/v1/model_registry/model-registry-service/registered_models?namespace=${profile_namespace}"
response_directory=$(mktemp -d)
control_probe="registry-egress-control-${RANDOM}"
policy_modified=false
sidecar_modified=false

cleanup() {
    local result=$?
    trap - EXIT
    if [[ $policy_modified == true ]]; then
        kubectl apply -f "$response_directory/registry-policy.json" || result=1
    fi
    if [[ $sidecar_modified == true ]]; then
        kubectl apply -f "$response_directory/sidecar.json" || result=1
    fi
    kubectl -n kubeflow delete pod "$control_probe" --ignore-not-found --wait=false || result=1
    rm -rf "$response_directory"
    exit "$result"
}
trap cleanup EXIT

# Remove server-managed fields so backups can also recreate deleted objects.
save_resource() {
    kubectl -n "$1" get "$2" "$3" -o json | python -c '
import json, sys
resource = json.load(sys.stdin)
resource.pop("status", None)
for field in ("resourceVersion", "uid", "creationTimestamp", "managedFields"):
    resource["metadata"].pop(field, None)
resource["metadata"].get("annotations", {}).pop("kubectl.kubernetes.io/last-applied-configuration", None)
json.dump(resource, sys.stdout)
' > "$4"
}
save_resource "$profile_namespace" authorizationpolicy model-registry-service "$response_directory/registry-policy.json"
save_resource kubeflow sidecar model-registry-ui-exception "$response_directory/sidecar.json"

# Replace the profile-local UI route with the real shared application. The
# existing gateway, identity propagation and per-user access reviews remain.
kustomize build tests/model_registry_shared_ui | kubectl apply -f -
kubectl -n "$profile_namespace" delete virtualservice model-registry-ui --ignore-not-found
kubectl -n kubeflow rollout status deployment/model-registry-ui --timeout=180s
kubectl -n kubeflow get deployment model-registry-ui -o json | python -c '
import json, sys
pod = json.load(sys.stdin)["spec"]["template"]["spec"]
assert pod["serviceAccountName"] == "model-registry-ui"
container = next(item for item in pod["containers"] if item["name"] == "model-registry-ui")
assert container["image"] == "ghcr.io/kubeflow/hub/ui:v0.3.17"
assert container["args"] == ["--deployment-mode=kubeflow", "--port=8080"]
assert not any(item["name"].startswith("MOCK") for item in container.get("env", []))
'

# Profile policies independently allow their own namespace's workloads. Use an
# unrelated control-plane identity rather than a pod in the target profile.
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${control_probe}
  namespace: kubeflow
  labels:
    app: registry-egress-control
spec:
  restartPolicy: Never
  serviceAccountName: default
  automountServiceAccountToken: false
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: registry-egress-probe
    image: busybox:1.37.0
    command: ["sleep", "900"]
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
kubectl -n kubeflow wait --for=condition=Ready "pod/$control_probe" --timeout=180s

# Do not enable shell tracing: the tokens must not appear in test logs.
profile_token=$(kubectl -n "$profile_namespace" create token default-editor --duration=10m)
unauthorized_token=$(kubectl -n default create token default --duration=10m)

assert_ui_response() {
    local expected_status=$1
    local token=$2
    local deadline=$((SECONDS + 90))
    local status
    while true; do
        status=$(curl --silent --show-error --max-time 20 \
            -o "$response_directory/ui.json" -w '%{http_code}' \
            -H "Authorization: Bearer $token" "$ui_route")
        if [[ $status == "$expected_status" ]]; then
            break
        fi
        if (( SECONDS >= deadline )); then
            printf 'FAIL: shared UI expected HTTP %s, received %s\n' "$expected_status" "$status" >&2
            return 1
        fi
        sleep 2
    done
    if [[ $expected_status == 200 ]]; then
        python -c 'import json, sys; data = json.load(open(sys.argv[1])); assert any(item["name"] == "test-model" for item in data["data"]["items"])' "$response_directory/ui.json"
    elif [[ $expected_status == 403 ]]; then
        python -c 'import json, sys; error = json.load(open(sys.argv[1]))["error"]; assert error == {"code": "403", "message": "Access forbidden"}' "$response_directory/ui.json"
    elif [[ $expected_status == 500 ]]; then
        python -c 'import json, sys; assert json.load(open(sys.argv[1]))["error"]["code"] == "500"' "$response_directory/ui.json"
    fi
    printf 'PASS: real shared UI returned HTTP %s\n' "$status"
}

assert_probe_denied() {
    local namespace=$1
    local probe=$2
    local endpoint=$3
    local response
    response=$(kubectl -n "$namespace" exec "$probe" -c registry-egress-probe -- \
        wget -S -O /dev/null -T 20 --header='kubeflow-userid: user@example.com' "$endpoint" 2>&1) || true
    grep -Eq 'HTTP/1[.][01] 403 ' <<< "$response" || {
        printf 'FAIL: unrelated workload %s/%s did not receive HTTP 403\n' "$namespace" "$probe" >&2
        return 1
    }
    printf 'PASS: unrelated workload %s/%s denied with HTTP 403\n' "$namespace" "$probe"
}

assert_ui_response 200 "$profile_token"
assert_ui_response 403 "$unauthorized_token"
status=$(curl --silent --show-error --max-time 20 -o /dev/null -w '%{http_code}' "$ui_route")
[[ $status == 302 || $status == 401 || $status == 403 ]] || {
    printf 'FAIL: unauthenticated UI request returned HTTP %s\n' "$status" >&2
    exit 1
}
printf 'PASS: unauthenticated gateway request denied with HTTP %s\n' "$status"
assert_probe_denied kubeflow "$control_probe" "$registry_endpoint"
assert_probe_denied kubeflow "$control_probe" "http://model-registry-ui-service.kubeflow.svc.cluster.local/api/v1/model_registry?namespace=${profile_namespace}"

# The exact additional trust rule must be necessary for the real UI request.
python -c 'import json, sys; rules = json.load(open(sys.argv[1]))["spec"]["rules"]; assert len(rules) == 3; assert rules[2] == {"from": [{"source": {"principals": ["cluster.local/ns/kubeflow/sa/model-registry-ui"]}}]}' "$response_directory/registry-policy.json"
policy_modified=true
kubectl -n "$profile_namespace" patch authorizationpolicy model-registry-service --type=json \
    -p='[{"op":"remove","path":"/spec/rules/2"}]'
assert_ui_response 500 "$profile_token"
kubectl apply -f "$response_directory/registry-policy.json"
assert_ui_response 200 "$profile_token"
policy_modified=false

# Prove the workload-specific egress exception is independently necessary.
sidecar_modified=true
kubectl -n kubeflow delete sidecar model-registry-ui-exception
assert_ui_response 500 "$profile_token"
kubectl apply -f "$response_directory/sidecar.json"
assert_ui_response 200 "$profile_token"
sidecar_modified=false
assert_ui_response 403 "$unauthorized_token"
printf 'Shared Registry UI cross-namespace authentication and authorization checks passed.\n'
