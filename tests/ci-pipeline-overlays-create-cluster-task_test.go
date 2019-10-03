package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writeCiPipelineOverlaysCreateClusterTask(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/config-map.yaml", `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: update-gcp-config
data:
  update-gcp-config.sh: |-
    #!/usr/bin/env bash
    sed 's/cluster-version:.*$/cluster-version: "$(cluster-version)"/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/cpu-pool-initialNodeCount:.*$/cpu-pool-initialNodeCount: $(cpu-pool-initialNodeCount)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/cpu-pool-machine-type:.*$/cpu-pool-machine-type: $(cpu-pool-machine-type)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/cpu-pool-max-nodes:.*$/cpu-pool-max-nodes: $(cpu-pool-max-nodes)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/cpu-pool-min-nodes:.*$/cpu-pool-min-nodes: $(cpu-pool-min-nodes)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed '18 a \    cpu-pool-min-cpu-platform: "$(cpu-pool-min-cpu-platform)"' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed '19 a \    cpu-pool-image-type: $(cpu-pool-image-type)' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/gpu-pool-initialNodeCount:.*$/gpu-pool-initialNodeCount: $(gpu-pool-initialNodeCount)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/gpu-pool-machine-type:.*$/gpu-pool-machine-type: $(gpu-pool-machine-type)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/gpu-pool-max-nodes:.*$/gpu-pool-max-nodes: $(gpu-pool-max-nodes)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/gpu-pool-min-nodes:.*$/gpu-pool-min-nodes: $(gpu-pool-min-nodes)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed 's/gpu-type:.*$/gpu-type: $(gpu-type)/' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed '43 a \    gpu-pool-min-cpu-platform: "$(gpu-pool-min-cpu-platform)"' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed '44 a \    gpu-pool-image-type: $(gpu-pool-image-type)' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
    sed "125 a \          imageType: {{ properties['cpu-pool-image-type'] }}" cluster.jinja > o
    mv o cluster.jinja
    sed "165 a \        imageType: {{ properties['gpu-pool-image-type'] }}" cluster.jinja > o
    mv o cluster.jinja
    sed "118, 131 s/minCpuPlatform: 'Intel Broadwell'/minCpuPlatform: {{ properties['cpu-pool-min-cpu-platform'] }}/" cluster.jinja > o
    mv o cluster.jinja
    sed "157, 175 s/minCpuPlatform: 'Intel Broadwell'/minCpuPlatform: {{ properties['gpu-pool-min-cpu-platform'] }}/" cluster.jinja > o
    mv o cluster.jinja
    sed 's/"\([0-9]*\)"/\1/g' cluster-kubeflow.yaml > o
    mv o cluster-kubeflow.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: label-accelerator-nodes
data:
  label-accelerator-nodes.sh: |-
    #!/usr/bin/env bash
    gpuNode=$(echo $(kubectl get nodes -oname --no-headers | grep gpu))
    if [[ -n $gpuNode ]]; then
      kubectl label $gpuNode accelerator=$(gpu-type)
    else
      echo 'no gpu node available'
    fi
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: label-cpu-nodes
data:
  label-cpu-nodes.sh: |-
    gcloud config set project $project
    cpuNodes=$(echo $(kubectl get nodes -oname --no-headers))
    if [[ -n $cpuNodes ]]; then
      for i in $cpuNodes;do
        cpuNode=${i#node/}
        cpuPlatform=$(gcloud compute instances describe $cpuNode --zone=us-central1-b --format json | jq .minCpuPlatform | xargs)
    echo 'cpuPlatform='$cpuPlatform
        case $cpuPlatform in
          'Intel Cascade Lake')
            kubectl label node $cpuNode cpu='cascadelake'
            ;;
          'Intel Skylake')
            kubectl label node $cpuNode cpu='skylake'
            ;;
          'Intel Broadwell')
            kubectl label node $cpuNode cpu='broadwell'
            ;;
          *)
            echo $cpuPlatform not supported
            ;;
        esac
      done
    else
      echo 'no cpu node available'
    fi
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/task.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: create-cluster
spec:
  inputs:
    params:
    - name: app_dir
      type: string
      description: where to create the kf app
    - name: cluster
      type: string
      description: name of the cluster
    - name: cluster-version
      type: string
      description: version of kubernetes
    - name: configPath
      type: string
      description: url for config arg
    - name: cpu-pool-initialNodeCount
      type: string
      description: initial cput node count
    - name: cpu-pool-machine-type
      type: string
      description: cpu machine type
    - name: cpu-pool-max-nodes
      type: string
      description: max cpu nodes
    - name: cpu-pool-min-nodes
      type: string
      description: min cpu nodes
    - name: cpu-pool-min-cpu-platform
      type: string
      description: cpu platform
    - name: cpu-pool-image-type
      type: string
      description: cpu image type
    - name: email
      type: string
      description: email for gcp
    - name: gpu-pool-initialNodeCount
      type: string
      description: initial gpu node count
    - name: gpu-pool-machine-type
      type: string
      description: gpu machine type
    - name: gpu-pool-max-nodes
      type: string
      description: max gpu nodes
    - name: gpu-pool-min-nodes
      type: string
      description: min gpu nodes
    - name: gpu-type
      type: string
      description: gpu type
    - name: gpu-pool-min-cpu-platform
      type: string
      description: cpu platform in gpu pool
    - name: gpu-pool-image-type
      type: string
      description: gpu image type
    - name: kfctl_image
      type: string
      description: the kfctl container image
    - name: namespace
      type: string
      description: the namespace to deploy kf 
    - name: platform
      type: string
      description: all | k8s
    - name: project
      type: string
      description: name of project
    - name: pvc_mount_path
      type: string
      description: parent dir for kfctl
    - name: zone
      type: string
      description: zone of project
  steps:
  - name: kfctl-activate-service-account
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "auth"
    - "activate-service-account"
    - "--key-file"
    - "/secret/gcp-credentials/key.json"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-set-account
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "config"
    - "set"
    - "account"
    - "$(inputs.params.email)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-init
    image: "$(inputs.params.kfctl_image)"
    workingDir: "$(inputs.params.pvc_mount_path)"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "init"
    - "--config"
    - "$(inputs.params.configPath)"
    - "--project"
    - "$(inputs.params.project)"
    - "--namespace"
    - "$(inputs.params.namespace)"
    - "$(inputs.params.app_dir)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow
      mountPath: "$(inputs.params.pvc_mount_path)"
    imagePullPolicy: IfNotPresent
  - name: kfctl-generate
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)/$(inputs.params.app_dir)"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "generate"
    - "$(inputs.params.platform)"
    - "--zone"
    - "$(inputs.params.zone)"
    - "--email"
    - "$(inputs.params.email)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: kubeflow-oauth
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: kubeflow-oauth
          key: CLIENT_SECRET
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow-oauth
      mountPath: /secret/kubeflow-oauth
    - name: kubeflow
      mountPath: /kubeflow
  - name: update-gcp-config
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)/$(inputs.params.app_dir)/gcp_config"
    #command: ["/bin/sleep", "infinity"]
    command: ["/bin/bash", "/update-gcp-config/update-gcp-config.sh"]
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow-oauth
      mountPath: /secret/kubeflow-oauth
    - name: kubeflow
      mountPath: /kubeflow
    - name: update-gcp-config
      mountPath: /update-gcp-config
  - name: kfctl-apply
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)/$(inputs.params.app_dir)"
    #command: ["/bin/sleep", "infinity"]
    command: ["/usr/local/bin/kfctl"]
    args:
    - "apply"
    - "$(inputs.params.platform)"
    - "--verbose"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: kubeflow-oauth
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: kubeflow-oauth
          key: CLIENT_SECRET
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow-oauth
      mountPath: /secret/kubeflow-oauth
    - name: kubeflow
      mountPath: /kubeflow
  - name: label-accelerator-nodes
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)"
    #command: ["/bin/sleep", "infinity"]
    command: ["/bin/bash", "/label-accelerator-nodes/label-accelerator-nodes.sh"]
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow-oauth
      mountPath: /secret/kubeflow-oauth
    - name: kubeflow
      mountPath: /kubeflow
    - name: label-accelerator-nodes
      mountPath: /label-accelerator-nodes
  - name: label-cpu-nodes
    image: "$(inputs.params.kfctl_image)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.pvc_mount_path)"
    #command: ["/bin/sleep", "infinity"]
    command: ["/bin/bash", "/label-cpu-nodes/label-cpu-nodes.sh"]
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/gcp-credentials/key.json
    envFrom:
    - configMapRef:
        name: ci-pipeline-run-parameters
    envFrom:
    - configMapRef:
        name: ci-pipeline-parameters
    volumeMounts:
    - name: gcp-credentials
      mountPath: /secret/gcp-credentials
    - name: kubeflow-oauth
      mountPath: /secret/kubeflow-oauth
    - name: kubeflow
      mountPath: /kubeflow
    - name: label-cpu-nodes
      mountPath: /label-cpu-nodes
  volumes:
  - name: gcp-credentials
    secret:
      secretName: gcp-credentials
  - name: kubeflow-oauth
    secret:
      secretName: kubeflow-oauth
  - name: kubeflow
    persistentVolumeClaim:
      claimName: ci-pipeline-run-persistent-volume-claim
  - name: update-gcp-config
    configMap:
      name: update-gcp-config
  - name: label-accelerator-nodes
    configMap:
      name: label-accelerator-nodes
  - name: label-cpu-nodes
    configMap:
      name: label-cpu-nodes
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/params.yaml", `
varReference:
- path: spec/params/value
  kind: Pipeline
- path: spec/tasks/params/value
  kind: Pipeline
- path: spec/resources/name
  kind: Pipeline
- path: spec/steps/image
  kind: Task
- path: spec/steps/volumeMounts/mountPath
  kind: Task
- path: spec/steps/workingDir
  kind: Task
- path: data/update-gcp-config.sh
  kind: ConfigMap
- path: data/label-accelerator-nodes.sh
  kind: ConfigMap
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/pipeline_patch.yaml", `
- op: add
  path: /spec/tasks/-
  value:
    name: create-cluster
    taskRef: 
      name: create-cluster
    params:
    - name: app_dir
      value: $(app_dir)
    - name: cluster
      value: $(cluster)
    - name: cluster-version
      value: $(cluster-version)
    - name: configPath
      value: $(configPath)
    - name: cpu-pool-initialNodeCount
      value: $(cpu-pool-initialNodeCount)
    - name: cpu-pool-initialNodeCount
      value: $(cpu-pool-initialNodeCount)
    - name: cpu-pool-machine-type
      value: $(cpu-pool-machine-type)
    - name: cpu-pool-max-nodes
      value: $(cpu-pool-max-nodes)
    - name: cpu-pool-min-nodes
      value: $(cpu-pool-min-nodes)
    - name: cpu-pool-min-cpu-platform
      value: $(cpu-pool-min-cpu-platform)
    - name: cpu-pool-image-type
      value: $(cpu-pool-image-type)
    - name: email
      value: $(email)
    - name: gpu-pool-initialNodeCount
      value: $(gpu-pool-initialNodeCount)
    - name: gpu-pool-machine-type
      value: $(gpu-pool-machine-type)
    - name: gpu-pool-max-nodes
      value: $(gpu-pool-max-nodes)
    - name: gpu-pool-min-nodes
      value: $(gpu-pool-min-nodes)
    - name: gpu-type
      value: $(gpu-type)
    - name: gpu-pool-min-cpu-platform
      value: $(gpu-pool-min-cpu-platform)
    - name: gpu-pool-image-type
      value: $(gpu-pool-image-type)
    - name: kfctl_image
      value: $(kfctl_image)
    - name: namespace
      value: $(namespace)
    - name: platform
      value: $(platform)
    - name: project
      value: $(project)
    - name: pvc_mount_path
      value: $(pvc_mount_path)
    - name: zone
      value: $(zone)
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/params.env", `
app_dir=/kubeflow/kubeflow-ci
cluster=kubeflow-ci
cluster-version=1.14
configPath=https://raw.githubusercontent.com/kubeflow/kubeflow/master/bootstrap/config/ci-cluster.yaml
cpu-pool-initialNodeCount=2
cpu-pool-machine-type=n1-standard-16
cpu-pool-max-nodes=10
cpu-pool-min-nodes=0
cpu-pool-min-cpu-platform=Intel Skylake
cpu-pool-image-type=ubuntu
email=foo@bar.com
gpu-pool-initialNodeCount=1
gpu-pool-machine-type=n1-standard-16
gpu-pool-max-nodes=1
gpu-pool-min-nodes=1
gpu-type=nvidia-tesla-p100
gpu-pool-min-cpu-platform=Intel Skylake
gpu-pool-image-type=ubuntu
kfctl_image=gcr.io/constant-cubist-173123/kfctl@sha256:dcdd79d565f936784021f463bf481930d21a797db140e94568977c4aaf5c018b
platform=all
project=constant-cubist-173123
pvc_mount_path=/kubeflow
zone=us-west1-a
`)
	th.writeK("/manifests/ci/ci-pipeline/overlays/create-cluster-task", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- config-map.yaml
- task.yaml
namespace: kubeflow-ci
patchesJson6902:
- target:
    group: tekton.dev
    version: v1alpha1
    kind: Pipeline
    name: ci-pipeline
  path: pipeline_patch.yaml
configMapGenerator:
- name: ci-pipeline-parameters
  behavior: merge
  env: params.env
vars:
- name: app_dir
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.app_dir
- name: cluster
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cluster
- name: cluster-version
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cluster-version
- name: configPath
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.configPath
- name: cpu-pool-initialNodeCount
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cpu-pool-initialNodeCount
- name: cpu-pool-machine-type
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cpu-pool-machine-type
- name: cpu-pool-max-nodes
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cpu-pool-max-nodes
- name: cpu-pool-min-nodes
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cpu-pool-min-nodes
- name: cpu-pool-min-cpu-platform
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cpu-pool-min-cpu-platform
- name: cpu-pool-image-type
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cpu-pool-image-type
- name: email
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.email
- name: gpu-pool-initialNodeCount
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-pool-initialNodeCount
- name: gpu-pool-machine-type
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-pool-machine-type
- name: gpu-pool-max-nodes
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-pool-max-nodes
- name: gpu-pool-min-nodes
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-pool-min-nodes
- name: gpu-type
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-type
- name: gpu-pool-min-cpu-platform
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-pool-min-cpu-platform
- name: gpu-pool-image-type
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gpu-pool-image-type
- name: kfctl_image
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.kfctl_image
- name: platform
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.platform
- name: pvc_mount_path
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pvc_mount_path
- name: project
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: zone
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.zone
configurations:
- params.yaml
`)
	th.writeF("/manifests/ci/ci-pipeline/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: ci-pipeline
  labels:
    scope: $(namespace)
spec:
  params: []
  resources: []
  tasks: []
`)
	th.writeF("/manifests/ci/ci-pipeline/base/params.env", `
namespace=
`)
	th.writeK("/manifests/ci/ci-pipeline/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
namespace: $(namespace)
configMapGenerator:
- name: ci-pipeline-parameters
  env: params.env
generatorOptions:
 disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
`)
}

func TestCiPipelineOverlaysCreateClusterTask(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline/overlays/create-cluster-task")
	writeCiPipelineOverlaysCreateClusterTask(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline/overlays/create-cluster-task"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
