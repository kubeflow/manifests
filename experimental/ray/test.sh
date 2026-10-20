#!/usr/bin/env bash
# The Helm workflow owns its operator release; this script owns its fixture.
set -euo pipefail
RAY_DIRECTORY=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
cd "$RAY_DIRECTORY"
NAMESPACE=${1:?Pass the existing Kubeflow Profile namespace}
RAY_MANAGED_BY_HELM=${RAY_MANAGED_BY_HELM:-false}
RAY_VERSION=2.56.0
PORT_FORWARD_PROCESS=""
FIXTURE_CREATED=false
OPERATOR_APPLIED=false
NAMESPACE_LABEL_ADDED=false
PROFILE_NAMESPACE_UID=""
PROBE_POD=ray-network-policy-probe
case "$RAY_MANAGED_BY_HELM" in
  true|false) ;;
  *) echo "RAY_MANAGED_BY_HELM must be true or false" >&2; exit 1 ;;
esac
# ShellCheck cannot follow the EXIT trap's function call here.
# shellcheck disable=SC2317,SC2329
cleanup() {
  local status=$?
  trap - EXIT
  if [[ -n "$PORT_FORWARD_PROCESS" ]]; then
    kill "$PORT_FORWARD_PROCESS" 2>/dev/null || true
    wait "$PORT_FORWARD_PROCESS" 2>/dev/null || true
  fi
  if [[ "$FIXTURE_CREATED" == true ]]; then
    kubectl -n "$NAMESPACE" delete pod "$PROBE_POD" --ignore-not-found --wait=true --timeout=120s || status=1
    kubectl -n "$NAMESPACE" delete -f raycluster_example.yaml --ignore-not-found --wait=true --timeout=120s || status=1
    local remaining=1
    for ((attempt=0; attempt<60; attempt++)); do
      remaining=$(kubectl -n "$NAMESPACE" get pods -l ray.io/cluster=kubeflow-raycluster -o json | jq '.items | length') || break
      [[ "$remaining" == 0 ]] && break
      sleep 2
    done
    if [[ "$remaining" != 0 ]]; then
      echo "Ray fixture Pods did not disappear" >&2
      status=1
    fi
  fi
  if [[ "$OPERATOR_APPLIED" == true ]]; then
    kustomize build kuberay-operator/overlays/kubeflow | kubectl -n kubeflow delete --ignore-not-found -f - || status=1
  fi
  if [[ "$NAMESPACE_LABEL_ADDED" == true ]]; then
    # Remove only the label added by this invocation, on the same Namespace.
    # The atomic tests refuse cleanup if either identity or label changed.
    kubectl patch namespace "$NAMESPACE" --type=json --patch \
      "[{\"op\":\"test\",\"path\":\"/metadata/uid\",\"value\":\"$PROFILE_NAMESPACE_UID\"},{\"op\":\"test\",\"path\":\"/metadata/labels/istio-injection\",\"value\":\"enabled\"},{\"op\":\"remove\",\"path\":\"/metadata/labels/istio-injection\"}]" \
      || status=1
  fi
  exit "$status"
}
trap cleanup EXIT
kubectl get namespace "$NAMESPACE" >/dev/null
# A temporary fixture must not overwrite existing Profile configuration.
ISTIO_INJECTION=$(kubectl get namespace "$NAMESPACE" -o jsonpath='{.metadata.labels.istio-injection}')
if [[ -n "$ISTIO_INJECTION" && "$ISTIO_INJECTION" != enabled ]]; then
  echo "The Profile namespace must allow istio-injection=enabled; refusing its existing value" >&2
  exit 1
fi
# Every object in the fixture must be absent before apply and cleanup can own it.
FIXTURE_RESOURCES=(
  authorizationpolicy.security.istio.io/allow-ray-workers-head
  service/raycluster-istio-headless-svc
  "pod/$PROBE_POD"
  # The operator refuses to adopt NetworkPolicies with these names.
  networkpolicy/kubeflow-raycluster-head
  networkpolicy/kubeflow-raycluster-workers-small-group
)
# The legacy path can install Ray for the first time. Without its definition,
# no RayCluster can exist; other read failures must stop the test.
RAY_DEFINITION=$(kubectl get crd rayclusters.ray.io --ignore-not-found -o name)
if [[ -n "$RAY_DEFINITION" ]]; then
  FIXTURE_RESOURCES+=(raycluster/kubeflow-raycluster)
fi
for fixture_resource in "${FIXTURE_RESOURCES[@]}"; do
  existing_fixture=$(kubectl -n "$NAMESPACE" get "$fixture_resource" --ignore-not-found -o name)
  if [[ -n "$existing_fixture" ]]; then
    echo "The test fixture $fixture_resource already exists" >&2
    exit 1
  fi
done
if [[ -z "$ISTIO_INJECTION" ]]; then
  PROFILE_NAMESPACE_UID=$(kubectl get namespace "$NAMESPACE" -o jsonpath='{.metadata.uid}')
  [[ -n "$PROFILE_NAMESPACE_UID" ]] || { echo "The Profile Namespace UID is missing" >&2; exit 1; }
  kubectl label namespace "$NAMESPACE" istio-injection=enabled
  NAMESPACE_LABEL_ADDED=true
fi
if [[ "$RAY_MANAGED_BY_HELM" == false ]]; then
  kustomize build kuberay-operator/overlays/kubeflow | kubectl -n kubeflow apply --server-side -f -
  OPERATOR_APPLIED=true
fi
kubectl -n kubeflow rollout status deployment/kuberay-operator --timeout=180s
FIXTURE_CREATED=true
kubectl -n "$NAMESPACE" apply -f raycluster_example.yaml
# The example contains Profile-specific routing references. Its supported
# namespace is documented; change the fixture explicitly for other Profiles.
for ((attempt=0; attempt<60; attempt++)); do
  if [[ $(kubectl -n "$NAMESPACE" get pods -l ray.io/cluster=kubeflow-raycluster -o json | jq '.items | length') -ge 2 ]]; then
    break
  fi
  sleep 3
done
kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l ray.io/cluster=kubeflow-raycluster --timeout=180s
HEAD_POD=$(kubectl -n "$NAMESPACE" get pods -l ray.io/cluster=kubeflow-raycluster,ray.io/node-type=head -o jsonpath='{.items[0].metadata.name}')
# Submit real distributed work and check its result, not only dashboard health.
run_distributed_tasks() {
  kubectl -n "$NAMESPACE" exec -i "$HEAD_POD" -c ray-head -- python - <<'PY'
import ray
from ray.util.scheduling_strategies import NodeAffinitySchedulingStrategy

ray.init(address="auto")

# The driver runs in the head Pod, so every other live node is a worker Pod.
head_node = ray.get_runtime_context().get_node_id()
worker_nodes = [
    node["NodeID"]
    for node in ray.nodes()
    if node["Alive"] and node["NodeID"] != head_node
]
assert worker_nodes, "No live Ray worker node is registered"


@ray.remote
def square(value):
    return value * value, ray.get_runtime_context().get_node_id()


# Hard node affinity forces every task onto a worker, so each task and its result
# cross the network between the head and worker Pods; a blocked path times out.
strategy = NodeAffinitySchedulingStrategy(node_id=worker_nodes[0], soft=False)
results = ray.get(
    [square.options(scheduling_strategy=strategy).remote(value) for value in range(10)],
    timeout=60,
)
assert [square for square, _ in results] == [value * value for value in range(10)], results
assert {node for _, node in results} == {worker_nodes[0]}, results
print("Ray distributed task results verified on worker node", worker_nodes[0])
ray.shutdown()
PY
}
run_distributed_tasks

# Verify denied access from an unmeshed probe and permitted access from a
# separate Istio-injected Ray Client while RayCluster isolation remains enabled.
HEAD_POD_IP=$(kubectl -n "$NAMESPACE" get pod "$HEAD_POD" -o jsonpath='{.status.podIP}')
[[ -n "$HEAD_POD_IP" ]] || { echo "The Ray head Pod has no IP address" >&2; exit 1; }
kubectl -n "$NAMESPACE" apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: $PROBE_POD
  annotations:
    sidecar.istio.io/inject: "false"
spec:
  restartPolicy: Never
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: probe
    image: rayproject/ray:${RAY_VERSION}-py311-cpu
    command: ["sleep", "infinity"]
    resources:
      requests:
        cpu: 10m
        memory: 64Mi
      limits:
        cpu: 100m
        memory: 256Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
EOF
kubectl -n "$NAMESPACE" wait --for=condition=Ready "pod/$PROBE_POD" --timeout=180s
probe_head() {
  local status=0
  kubectl -n "$NAMESPACE" exec "$PROBE_POD" -c probe -- python -c '
import socket
import sys

try:
  socket.create_connection((sys.argv[1], 6379), timeout=5).close()
except OSError:
    sys.exit(3)
' "$HEAD_POD_IP" || status=$?
  case "$status" in
    0) echo reachable ;;
    3) echo refused ;;
    *) echo "The network policy probe failed with status $status" >&2; return 1 ;;
  esac
}
wait_for_probe_result() {
  local expected=$1
  local result=""
  for ((attempt=0; attempt<20; attempt++)); do
    result=$(probe_head)
    [[ "$result" == "$expected" ]] && return 0
    sleep 3
  done
  echo "The Ray head stayed $result for the client; expected $expected" >&2
  return 1
}
wait_for_probe_result reachable
kubectl -n "$NAMESPACE" patch raycluster kubeflow-raycluster --type=merge \
  --patch '{"spec":{"networkPolicy":{"mode":"DenyAllIngress"}}}'
for network_policy in kubeflow-raycluster-head kubeflow-raycluster-workers-small-group; do
  kubectl -n "$NAMESPACE" wait --for=create "networkpolicy/$network_policy" --timeout=120s
  kubectl -n "$NAMESPACE" get networkpolicy "$network_policy" -o json \
    | jq -e '.spec.policyTypes == ["Ingress"] and .metadata.ownerReferences[0].kind == "RayCluster"' >/dev/null
done
wait_for_probe_result refused
CLIENT_POD="${PROBE_POD}-client"
kubectl -n "$NAMESPACE" get pod "$PROBE_POD" -o json \
  | jq '{
      apiVersion,
      kind,
      metadata: {
        name: (.metadata.name + "-client"),
        labels: {"ray-network-policy-client": "true"},
        annotations: {"sidecar.istio.io/inject": "true"},
        ownerReferences: [{apiVersion, kind, name: .metadata.name, uid: .metadata.uid}]
      },
      spec: {
        serviceAccountName: "default-editor",
        restartPolicy: .spec.restartPolicy,
        securityContext: .spec.securityContext,
        containers: [.spec.containers[] | {
          name, image, command, resources, securityContext
        } | .resources.limits = {cpu: "500m", memory: "512Mi"}]
      }
    }' \
  | kubectl -n "$NAMESPACE" create -f -
kubectl -n "$NAMESPACE" wait --for=condition=Ready "pod/$CLIENT_POD" --timeout=180s
kubectl -n "$NAMESPACE" patch raycluster kubeflow-raycluster --type=merge \
  --patch "$(jq -nc '{
    spec: {networkPolicy: {head: {ingressRules: [{
      from: [{podSelector: {matchLabels: {"ray-network-policy-client": "true"}}}],
      ports: [{protocol: "TCP", port: 10001}]
    }]}}}
  }')"
timeout 120s kubectl -n "$NAMESPACE" exec -i "$CLIENT_POD" -c probe \
  -- python - <<'PY'
import ray

ray.init(address="ray://kubeflow-raycluster-head-svc:10001")

@ray.remote
def square(value):
    return value * value

result = ray.get([square.remote(value) for value in range(10)], timeout=60)
assert result == [value * value for value in range(10)], result
print("Ray Client task results verified:", result)
ray.shutdown()
PY
wait_for_probe_result refused
run_distributed_tasks
echo "RayClusterNetworkPolicy isolation and permitted Ray Client access verified"

kubectl -n "$NAMESPACE" port-forward svc/kubeflow-raycluster-head-svc 8265:8265 &
PORT_FORWARD_PROCESS=$!
for ((attempt=0; attempt<30; attempt++)); do
  if output=$(curl --fail --silent --show-error --max-time 3 localhost:8265/api/version); then
    echo "$output" | jq -e --arg version "$RAY_VERSION" '.ray_version == $version'
    exit 0
  fi
  sleep 2
done
echo "Ray dashboard did not become reachable" >&2
exit 1
