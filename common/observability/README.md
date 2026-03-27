### Overview
An opt-in kustomize component providing a complete monitoring foundation for Kubeflow clusters with GPU workloads. Installs Prometheus Operator, Grafana Operator, GPU ServiceMonitors for NVIDIA DCGM and AMD ROCm, and three Grafana dashboards. Kepler energy metrics are available as a separate opt-in sub-component.

### Prerequisites
| Prerequisite | Required for | Notes |
|---|---|---|
| Kubernetes 1.27+ | Everything | |
| kustomize v5+ | Installation | |
| NVIDIA GPU Operator | NVIDIA ServiceMonitor | Deploy in `gpu-operator` ns — silent if absent |
| AMD GPU Operator | AMD ServiceMonitor | Deploy in `kube-amd-gpu` ns — silent if absent |
| kube-state-metrics | GPU Namespace Usage + Availability dashboards | **Without it 2/3 dashboards render blank with no error** — install via kube-prometheus-stack or standalone |

### Installation
```sh
# Main stack (Prometheus + Grafana + ServiceMonitors + dashboards)
kustomize build common/observability/overlays/kubeflow | kubectl apply --server-side -f -

# Or via script
./tests/observability_install.sh

# Kepler energy metrics (opt-in — separate step)
kustomize build common/observability/components/kepler | kubectl apply --server-side -f -
```
Note: `--server-side` is required — CRD bundles exceed client-side annotation size limits.

### What gets installed
| Resource | Namespace | Purpose |
|---|---|---|
| Prometheus Operator | kubeflow-monitoring-system | Manages Prometheus CR |
| Prometheus CR | kubeflow-monitoring-system | Scrapes all ServiceMonitors across all namespaces |
| Grafana Operator | kubeflow-monitoring-system | Manages Grafana, GrafanaDatasource, GrafanaDashboard CRs |
| Grafana CR | kubeflow-monitoring-system | Grafana instance |
| GrafanaDatasource | kubeflow-monitoring-system | Prometheus datasource, uid: prometheus |
| NVIDIA DCGM ServiceMonitor | gpu-operator | Scrapes DCGM exporter |
| AMD ROCm ServiceMonitor | kube-amd-gpu | Scrapes device-metrics-exporter |
| 3x GrafanaDashboard CRs | kubeflow-monitoring-system | GPU dashboards |
| Kepler DaemonSet (opt-in) | kepler | Per-pod energy/power draw metrics |
| Kepler ServiceMonitor (opt-in) | kubeflow-monitoring-system | Scrapes Kepler |

### Dashboards
| Dashboard | What it shows | Dependencies |
|---|---|---|
| GPU Cluster Usage | Cluster-wide GPU utilization, memory, count per node | DCGM or ROCm metrics |
| GPU Namespace Usage | Per-namespace GPU allocation and utilization | kube-state-metrics + DCGM |
| GPU Availability & Allocation | Allocation ratios, pending GPU sessions | kube-state-metrics |

### Accessing Grafana
```sh
kubectl port-forward svc/grafana-service -n kubeflow-monitoring-system 3000:3000
```
Open http://localhost:3000 — default credentials are admin/admin on first login.

### Kepler (opt-in)
Kepler deploys to its own `kepler` namespace (PSS: privileged) to avoid impacting the PSS restricted posture of `kubeflow-monitoring-system`. It requires `privileged: true` to access `/proc`, `/sys`, and the container runtime socket for energy metrics.

### Reference
- CERN architecture: https://architecture.cncf.io/architectures/cern-scientific-computing/
- Issue: https://github.com/kubeflow/manifests/issues/3426
