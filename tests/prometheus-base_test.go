package tests_test

import (
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"testing"
)

func writePrometheusBase(th *KustTestHarness) {
	th.writeF("/manifests/gcp/prometheus/base/prometheus.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    ksonnet.io/component: prometheus
  name: stackdriver
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    ksonnet.io/component: prometheus
  name: prometheus
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
      - nodes/proxy
      - services
      - endpoints
      - pods
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
  - nonResourceURLs:
      - /metrics
    verbs:
      - get
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    ksonnet.io/component: prometheus
  name: prometheus-stackdriver
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
  - kind: ServiceAccount
    name: prometheus
    namespace: stackdriver
---
apiVersion: v1
data:
  prometheus.yml: |
    # Source: https://github.com/stackdriver/prometheus/blob/master/documentation/examples/prometheus.yml
    global:
      external_labels:
        _stackdriver_project_id: $(projectId)
        _kubernetes_cluster_name: $(clusterName)
        _kubernetes_location: $(zone)

    # Scrape config for nodes (kubelet).
    #
    # Rather than connecting directly to the node, the scrape is proxied though the
    # Kubernetes apiserver.  This means it will work if Prometheus is running out of
    # cluster, or can't connect to nodes for some other reason (e.g. because of
    # firewalling).
    scrape_configs:
    - job_name: 'kubernetes-nodes'

      # Default to scraping over https. If required, just disable this or change to
      # http
      scheme: https

      # This TLS & bearer token file config is used to connect to the actual scrape
      # endpoints for cluster components. This is separate to discovery auth
      # configuration because discovery & scraping are two separate concerns in
      # Prometheus. The discovery auth config is automatic if Prometheus runs inside
      # the cluster. Otherwise, more config options have to be provided within the
      # <kubernetes_sd_config>.
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

      kubernetes_sd_configs:
      - role: node

      relabel_configs:
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics

    # Example scrape config for pods
    #
    # The relabeling allows the actual pod scrape endpoint to be configured via the
    # following annotations:
    #
    # * "prometheus.io/scrape": Only scrape pods that have a value of "true"
    # * "prometheus.io/path": If the metrics path is not "/metrics" override this.
    # * "prometheus.io/port": Scrape the pod on the indicated port instead of the
    # pod's declared ports (default is a port-free target if none are declared).
    - job_name: 'kubernetes-pods-containers'

      kubernetes_sd_configs:
      - role: pod

      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__

    # Scrape config for service endpoints.
    #
    # The relabeling allows the actual service scrape endpoint to be configured
    # via the following annotations:
    #
    # * "prometheus.io/scrape": Only scrape services that have a value of "true"
    # * "prometheus.io/scheme": If the metrics endpoint is secured then you will need
    # to set this to "https" & most likely set the "tls_config" of the scrape config.
    # * "prometheus.io/path": If the metrics path is not "/metrics" override this.
    # * "prometheus.io/port": If the metrics are exposed on a different port to the
    # service then set this appropriately.
    - job_name: 'kubernetes-service-endpoints'

      kubernetes_sd_configs:
      - role: endpoints

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2


    # Scrape config for k8s services
    - job_name: 'kubernetes-services'

      kubernetes_sd_configs:
      - role: service

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
      - source_labels: [__address__,__meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: (.+)(?::\d+);(\d+)
        replacement: $1:$2

    remote_write:
    - url: "https://monitoring.googleapis.com:443/"
      queue_config:
        # Capacity should be 2*max_samples_per_send.
        capacity: 2000
        max_samples_per_send: 1000
        max_shards: 10000
      write_relabel_configs:
      # These labels are generally redundant with the Stackdriver monitored resource labels.
      - source_labels: [job]
        target_label: job
        replacement: ""
      - source_labels: [instance]
        target_label: instance
        replacement: ""
kind: ConfigMap
metadata:
  labels:
    ksonnet.io/component: prometheus
  name: prometheus
  namespace: stackdriver
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    ksonnet.io/component: prometheus
  name: prometheus
  namespace: stackdriver
---
apiVersion: v1
kind: Service
metadata:
  labels:
    ksonnet.io/component: prometheus
    name: prometheus
  name: prometheus
  namespace: stackdriver
spec:
  ports:
    - name: prometheus
      port: 9090
      protocol: TCP
  selector:
    app: prometheus
  type: ClusterIP
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    ksonnet.io/component: prometheus
  name: prometheus
  namespace: stackdriver
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
      labels:
        app: prometheus
      name: prometheus
      namespace: stackdriver
    spec:
      containers:
        - image: gcr.io/stackdriver-prometheus/stackdriver-prometheus:release-0.4.2
          imagePullPolicy: Always
          name: prometheus
          ports:
            - containerPort: 9090
              name: web
          resources:
            limits:
              cpu: 400m
              memory: 1000Mi
            requests:
              cpu: 20m
              memory: 50Mi
          volumeMounts:
            - mountPath: /etc/prometheus
              name: config-volume
      serviceAccountName: prometheus
      volumes:
        - configMap:
            name: prometheus
          name: config-volume
`)
	th.writeF("/manifests/gcp/prometheus/base/params.yaml", `
varReference:
- path: data/prometheus.yml
  kind: ConfigMap

`)
	th.writeF("/manifests/gcp/prometheus/base/params.env", `
projectId=
clusterName=
zone=
`)
	th.writeK("/manifests/gcp/prometheus/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- prometheus.yaml
commonLabels:
  kustomize.component: prometheus
configMapGenerator:
  - name: prometheus-parameters
    env: params.env
images:
  - name: gcr.io/stackdriver-prometheus/stackdriver-prometheus
    newName: gcr.io/stackdriver-prometheus/stackdriver-prometheus
    newTag: release-0.4.2
vars:
  - name: projectId
    objref:
      kind: ConfigMap
      name: prometheus-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.projectId
  - name: clusterName
    objref:
      kind: ConfigMap
      name: prometheus-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.clusterName
  - name: zone
    objref:
      kind: ConfigMap
      name: prometheus-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.zone
configurations:
- params.yaml
`)
}

func TestPrometheusBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/prometheus/base")
	writePrometheusBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/prometheus/base"
	fsys := fs.MakeRealFS()
	_loader, loaderErr := loader.NewLoader(targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	n, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := n.EncodeAsYaml()
	th.assertActualEqualsExpected(m, string(expected))
}
