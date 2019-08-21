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

func writeTektoncdInstallBase(th *KustTestHarness) {
	th.writeF("/manifests/tektoncd/tektoncd-install/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: tekton-pipelines-controller-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-pipelines-admin
subjects:
- kind: ServiceAccount
  name: tekton-pipelines-controller
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: tekton-pipelines-admin
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - namespaces
  - secrets
  - events
  - serviceaccounts
  - configmaps
  - persistentvolumeclaims
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments/finalizers
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - tekton.dev
  resources:
  - tasks
  - clustertasks
  - taskruns
  - pipelines
  - pipelineruns
  - pipelineresources
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - tekton.dev
  resources:
  - taskruns/finalizers
  - pipelineruns/finalizers
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - tekton.dev
  resources:
  - tasks/status
  - clustertasks/status
  - taskruns/status
  - pipelines/status
  - pipelineruns/status
  - pipelineresources/status
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - policy
  resourceNames:
  - tekton-pipelines
  resources:
  - podsecuritypolicies
  verbs:
  - use
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
  name: tekton-aggregate-edit
rules:
- apiGroups:
  - tekton.dev
  resources:
  - tasks
  - taskruns
  - pipelines
  - pipelineruns
  - pipelineresources
  verbs:
  - create
  - delete
  - deletecollection
  - get
  - list
  - patch
  - update
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    rbac.authorization.k8s.io/aggregate-to-view: "true"
  name: tekton-aggregate-view
rules:
- apiGroups:
  - tekton.dev
  resources:
  - tasks
  - taskruns
  - pipelines
  - pipelineruns
  - pipelineresources
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/config-map.yaml", `
---
apiVersion: v1
data: null
kind: ConfigMap
metadata:
  name: config-artifact-bucket
---
apiVersion: v1
data: null
kind: ConfigMap
metadata:
  name: config-artifact-pvc
---
apiVersion: v1
data:
  _example: |
    ################################
    #                              #
    #    EXAMPLE CONFIGURATION     #
    #                              #
    ################################

    # This block is not actually functional configuration,
    # but serves to illustrate the available configuration
    # options and document them in a way that is accessible
    # to users that 'kubectl edit' this config map.
    #
    # These sample configuration options may be copied out of
    # this example block and unindented to be in the data block
    # to actually change the configuration.

    # default-timeout-minutes contains the default number of
    # minutes to use for TaskRun and PipelineRun, if none is specified.
    default-timeout-minutes: "60"  # 60 minutes
kind: ConfigMap
metadata:
  name: config-defaults
---
apiVersion: v1
data:
  _example: |
    ################################
    #                              #
    #    EXAMPLE CONFIGURATION     #
    #                              #
    ################################

    # This block is not actually functional configuration,
    # but serves to illustrate the available configuration
    # options and document them in a way that is accessible
    # to users that 'kubectl edit' this config map.
    #
    # These sample configuration options may be copied out of
    # this example block and unindented to be in the data block
    # to actually change the configuration.

    # metrics.backend-destination field specifies the system metrics destination.
    # It supports either prometheus (the default) or stackdriver.
    # Note: Using stackdriver will incur additional charges
    metrics.backend-destination: prometheus

    # metrics.stackdriver-project-id field specifies the stackdriver project ID. This
    # field is optional. When running on GCE, application default credentials will be
    # used if this field is not provided.
    metrics.stackdriver-project-id: "<your stackdriver project id>"

    # metrics.allow-stackdriver-custom-metrics indicates whether it is allowed to send metrics to
    # Stackdriver using "global" resource type and custom metric type if the
    # metrics are not supported by "knative_revision" resource type. Setting this
    # flag to "true" could cause extra Stackdriver charge.
    # If metrics.backend-destination is not Stackdriver, this is ignored.
    metrics.allow-stackdriver-custom-metrics: "false"
kind: ConfigMap
metadata:
  name: config-observability
---
apiVersion: v1
data:
  loglevel.controller: info
  loglevel.webhook: info
  zap-logger-config: |
    {
      "level": "info",
      "development": false,
      "sampling": {
        "initial": 100,
        "thereafter": 100
      },
      "outputPaths": ["stdout"],
      "errorOutputPaths": ["stderr"],
      "encoding": "json",
      "encoderConfig": {
        "timeKey": "",
        "levelKey": "level",
        "nameKey": "logger",
        "callerKey": "caller",
        "messageKey": "msg",
        "stacktraceKey": "stacktrace",
        "lineEnding": "",
        "levelEncoder": "",
        "timeEncoder": "",
        "durationEncoder": "",
        "callerEncoder": ""
      }
    }
kind: ConfigMap
metadata:
  name: config-logging
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/deployment.yaml", `
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: tekton-pipelines-controller
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
      labels:
        app: tekton-pipelines-controller
    spec:
      containers:
      - args:
        - -logtostderr
        - -stderrthreshold
        - INFO
        - -kubeconfig-writer-image
        - $(registry)/$(kubeconfigwriter)
        - -creds-image
        - $(registry)/$(creds-init)
        - -git-image
        - $(registry)/$(git-init)
        - -nop-image
        - $(registry)/$(nop)
        - -bash-noop-image
        - $(registry)/$(bash)
        - -gsutil-image
        - $(registry)/$(gsutil)
        - -entrypoint-image
        - $(registry)/$(entrypoint)
        - -imagedigest-exporter-image
        - $(registry)/$(imagedigestexporter)
        - -pr-image
        - $(registry)/$(pullrequest-init)
        - -build-gcs-fetcher-image
        - $(registry)/$(gcs-fetcher)
        - -build-gcs-uploader-image
        - $(registry)/$(gcs-uploader)
        image: $(registry)/$(controller)
        name: tekton-pipelines-controller
        volumeMounts:
        - mountPath: /etc/config-logging
          name: config-logging
      serviceAccountName: tekton-pipelines-controller
      volumes:
      - configMap:
          name: config-logging
        name: config-logging
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: tekton-pipelines-webhook
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
      labels:
        app: tekton-pipelines-webhook
    spec:
      containers:
      - name: webhook
        image: $(registry)/$(webhook)
        volumeMounts:
        - mountPath: /etc/config-logging
          name: config-logging
      serviceAccountName: tekton-pipelines-controller
      volumes:
      - configMap:
          name: config-logging
        name: config-logging
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/pod-security-policy.yaml", `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: tekton-pipelines
spec:
  allowPrivilegeEscalation: false
  fsGroup:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  hostIPC: false
  hostNetwork: false
  hostPID: false
  privileged: false
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  volumes:
  - emptyDir
  - configMap
  - secret
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tekton-pipelines-controller
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: tekton-pipelines-controller
  name: tekton-pipelines-controller
spec:
  ports:
  - name: metrics
    port: 9090
    protocol: TCP
    targetPort: 9090
  selector:
    app: tekton-pipelines-controller
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: tekton-pipelines-webhook
  name: tekton-pipelines-webhook
spec:
  ports:
  - port: 443
    targetPort: 8443
  selector:
    app: tekton-pipelines-webhook
---
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/image
  kind: Deployment
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/params.env", `
clusterDomain=cluster.local
project=constant-cubist-173123
registry=gcr.io/tekton-releases
webhook=github.com/tektoncd/pipeline/cmd/webhook@sha256:496e36b8723a668ac3531acc26512c123342da7827c10386b571aa975d6a47e7
nop=github.com/tektoncd/pipeline/cmd/nop@sha256:c903f9e4d60220e7cf7beab4b94e4117abcc048ab7404da3a2a4b417891741cb
entrypoint=github.com/tektoncd/pipeline/cmd/entrypoint@sha256:a424ab773b89e13e5e03ff90962db98424621b47c1bb543ec270783cfd859faf
gsutil=github.com/tektoncd/pipeline/cmd/gsutil@sha256:8a86ac637e78885d2945025b43da950a0058f36b3dc62c2bc623963ace19ca1b
gcs-fetcher=github.com/tektoncd/pipeline/vendor/github.com/googlecloudplatform/cloud-builders/gcs-fetcher/cmd/gcs-uploader@sha256:2ac1b03e3a262511d6bfdc7c2d05db517dc1c785830f383f132b8d97dba22947
gcs-uploader=github.com/tektoncd/pipeline/vendor/github.com/googlecloudplatform/cloud-builders/gcs-fetcher/cmd/gcs-uploader@sha256:2ac1b03e3a262511d6bfdc7c2d05db517dc1c785830f383f132b8d97dba22947
bash=github.com/tektoncd/pipeline/cmd/bash@sha256:157b21c4b29a4f2aa96d52add55781f211cc8101df36657b82089119b2fc4004
creds-init=github.com/tektoncd/pipeline/cmd/creds-init@sha256:c0235af1723068e6806def1d998436cde5d93ff1c38a94b9c92410f5f01bcb26
git-init=github.com/tektoncd/pipeline/cmd/git-init@sha256:2e5217266f515f91be333d5f8abcdc98bb1a7a4de7b339734e10fd7b972eeb5f
pullrequest-init=github.com/tektoncd/pipeline/cmd/pullrequest-init@sha256:da5dfe24ae824e5e737cee57b2a248eee15e128b0cca44f9466bab902fa8bea0
imagedigestexporter=github.com/tektoncd/pipeline/cmd/imagedigestexporter@sha256:aae9c44ed56f0d30530a2349f255c4977a6d8d4a497dfdca626b51f35bf229b4
kubeconfigwriter=github.com/tektoncd/pipeline/cmd/kubeconfigwriter@sha256:115acf8aa4d79be49a481f6d520ff66839d57656c840588052097956224fb3ff
controller=github.com/tektoncd/pipeline/cmd/controller@sha256:4f10413791df045f29f882fab817219e54123b527d6230a4991e2558f3d659f9
`)
	th.writeK("/manifests/tektoncd/tektoncd-install/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- deployment.yaml
- pod-security-policy.yaml
- service-account.yaml
- service.yaml
namespace: tekton-pipelines
configMapGenerator:
- name: tektoncd-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: project
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: registry
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.registry
- name: entrypoint
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.entrypoint
- name: nop
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.nop
- name: webhook
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.webhook
- name: gcs-fetcher
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gcs-fetcher
- name: gcs-uploader
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gcs-uploader
- name: gsutil
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gsutil
- name: bash
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.bash
- name: git-init
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.git-init
- name: creds-init
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.creds-init
- name: pullrequest-init
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pullrequest-init
- name: imagedigestexporter
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.imagedigestexporter
- name: kubeconfigwriter
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.kubeconfigwriter
- name: controller
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.controller
configurations:
- params.yaml
`)
}

func TestTektoncdInstallBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tektoncd/tektoncd-install/base")
	writeTektoncdInstallBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tektoncd/tektoncd-install/base"
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
