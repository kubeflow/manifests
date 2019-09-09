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

func writeCiPipelineOverlaysCreateClusterTask(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: pipeline
spec:
  params:
  - name: namespace
    value: $(namespace)
  - name: app_dir
    value: $(app_dir)
  - name: project
    value: $(project)
  - name: configPath
    value: $(configPath)
  - name: zone
    value: $(zone)
  - name: email
    value: $(email)
  - name: platform
    value: $(platform)
  - name: cluster
    value: $(cluster)
  - name: kfctl_image
    value: $(kfctl_image)
  - name: pvc_mount_path
    value: $(pvc_mount_path)
  tasks:
  - name: create-cluster-task
    taskRef: 
      name: create-cluster-task
    params:
    - name: kfctl_image
      type: string
      description: the kfctl container image
    - name: namespace
      type: string
      description: the namespace to deploy kf 
    - name: app_dir
      type: string
      description: where to create the kf app
    - name: configPath
      type: string
      description: url for config arg
    - name: project
      type: string
      description: name of project
    - name: zone
      type: string
      description: zone of project
    - name: platform
      type: string
      description: all | k8s
    - name: email
      type: string
      description: email for gcp
    - name: cluster
      type: string
      description: name of the cluster
    - name: pvc_mount_path
      type: string
      description: parent dir for kfctl
    steps:
    - name: kfctl-activate-service-account
      image: "${inputs.params.kfctl_image}"
      imagePullPolicy: IfNotPresent
      workingDir: "${inputs.params.pvc_mount_path}"
      command: ["/opt/google-cloud-sdk/bin/gcloud"]
      args:
      - "auth"
      - "activate-service-account"
      - "--key-file"
      - "/secret/create-cluster-gcp-secret.json"
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/create-cluster-gcp-secret.json
      - name: CLIENT_ID
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_ID
      - name: CLIENT_SECRET
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_SECRET
      volumeMounts:
      - name: create-cluster-gcp-secret
        mountPath: /secret
      - name: kubeflow
        mountPath: /kubeflow
    - name: kfctl-set-account
      image: "${inputs.params.kfctl_image}"
      imagePullPolicy: IfNotPresent
      workingDir: "${inputs.params.pvc_mount_path}"
      command: ["/opt/google-cloud-sdk/bin/gcloud"]
      args:
      - "config"
      - "set"
      - "account"
      - "${inputs.params.email}"
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/create-cluster-gcp-secret.json
      - name: CLIENT_ID
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_ID
      - name: CLIENT_SECRET
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_SECRET
      volumeMounts:
      - name: create-cluster-gcp-secret
        mountPath: /secret
      - name: kubeflow
        mountPath: /kubeflow
    - name: kfctl-init
      image: "${inputs.params.kfctl_image}"
      workingDir: "${inputs.params.pvc_mount_path}"
      command: ["/usr/local/bin/kfctl"]
      args:
      - "init"
      - "--config"
      - "${inputs.params.configPath}"
      - "--project"
      - "${inputs.params.project}"
      - "--namespace"
      - "${inputs.params.namespace}"
      - "${inputs.params.app_dir}"
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/create-cluster-kaniko-secret.json
      volumeMounts:
      - name: create-cluster-kaniko-secret
        mountPath: /secret
      - name: kubeflow
        mountPath: "${inputs.params.pvc_mount_path}"
      imagePullPolicy: IfNotPresent
    - name: kfctl-generate
      image: "${inputs.params.kfctl_image}"
      imagePullPolicy: IfNotPresent
      workingDir: "${inputs.params.pvc_mount_path}/${inputs.params.app_dir}"
      command: ["/usr/local/bin/kfctl"]
      args:
      - "generate"
      - "${inputs.params.platform}"
      - "--zone"
      - "${inputs.params.zone}"
      - "--email"
      - "${inputs.params.email}"
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/create-cluster-kaniko-secret.json
      - name: CLIENT_ID
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_ID
      - name: CLIENT_SECRET
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_SECRET
      volumeMounts:
      - name: create-cluster-kaniko-secret
        mountPath: /secret
      - name: kubeflow
        mountPath: /kubeflow
    - name: kfctl-apply
      image: "${inputs.params.kfctl_image}"
      imagePullPolicy: IfNotPresent
      workingDir: "${inputs.params.pvc_mount_path}/${inputs.params.app_dir}"
  #    command: ["/bin/sleep", "infinity"]
      command: ["/usr/local/bin/kfctl"]
      args:
      - "apply"
      - "${inputs.params.platform}"
      - "--verbose"
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/create-cluster-gcp-secret.json
      - name: CLIENT_ID
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_ID
      - name: CLIENT_SECRET
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_SECRET
      volumeMounts:
      - name: create-cluster-gcp-secret
        mountPath: /secret
      - name: kubeflow
        mountPath: /kubeflow
    - name: kfctl-configure-kubectl
      image: "${inputs.params.kfctl_image}"
      imagePullPolicy: IfNotPresent
      workingDir: "${inputs.params.pvc_mount_path}"
      command: ["/opt/google-cloud-sdk/bin/gcloud"]
      args:
      - "--project"
      - "${inputs.params.project}"
      - "container"
      - "clusters"
      - "--zone"
      - "${inputs.params.zone}"
      - "get-credentials"
      - "${inputs.params.cluster}"
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/create-cluster-gcp-secret.json
      - name: CLIENT_ID
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_ID
      - name: CLIENT_SECRET
        valueFrom:
          secretKeyRef:
            name: create-cluster-client-secret
            key: CLIENT_SECRET
      volumeMounts:
      - name: create-cluster-gcp-secret
        mountPath: /secret
      - name: kubeflow
        mountPath: /kubeflow
    volumes:
    - name: create-cluster-kaniko-secret
      secret:
        secretName: create-cluster-kaniko-secret
    - name: create-cluster-docker-secret
      secret:
        secretName: create-cluster-docker-secret
    - name: create-cluster-gcp-secret
      secret:
        secretName: create-cluster-gcp-secret
    - name: kubeflow
      persistentVolumeClaim:
        claimName: create-cluster-persistent-volume-claim
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/task.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: task
spec:
  inputs:
    params:
    - name: kfctl_image
      type: string
      description: the kfctl container image
    - name: namespace
      type: string
      description: the namespace to deploy kf 
    - name: app_dir
      type: string
      description: where to create the kf app
    - name: configPath
      type: string
      description: url for config arg
    - name: project
      type: string
      description: name of project
    - name: zone
      type: string
      description: zone of project
    - name: platform
      type: string
      description: all | k8s
    - name: email
      type: string
      description: email for gcp
    - name: cluster
      type: string
      description: name of the cluster
    - name: pvc_mount_path
      type: string
      description: parent dir for kfctl
  steps:
  - name: kfctl-activate-service-account
    image: "${inputs.params.kfctl_image}"
    imagePullPolicy: IfNotPresent
    workingDir: "${inputs.params.pvc_mount_path}"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "auth"
    - "activate-service-account"
    - "--key-file"
    - "/secret/ci-create-cluster-gcp-secret.json"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-create-cluster-gcp-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-create-cluster-gcp-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-set-account
    image: "${inputs.params.kfctl_image}"
    imagePullPolicy: IfNotPresent
    workingDir: "${inputs.params.pvc_mount_path}"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "config"
    - "set"
    - "account"
    - "${inputs.params.email}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-create-cluster-gcp-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-create-cluster-gcp-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-init
    image: "${inputs.params.kfctl_image}"
    workingDir: "${inputs.params.pvc_mount_path}"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "init"
    - "--config"
    - "${inputs.params.configPath}"
    - "--project"
    - "${inputs.params.project}"
    - "--namespace"
    - "${inputs.params.namespace}"
    - "${inputs.params.app_dir}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-create-cluster-kaniko-secret.json
    volumeMounts:
    - name: ci-create-cluster-kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: "${inputs.params.pvc_mount_path}"
    imagePullPolicy: IfNotPresent
  - name: kfctl-generate
    image: "${inputs.params.kfctl_image}"
    imagePullPolicy: IfNotPresent
    workingDir: "${inputs.params.pvc_mount_path}/${inputs.params.app_dir}"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "generate"
    - "${inputs.params.platform}"
    - "--zone"
    - "${inputs.params.zone}"
    - "--email"
    - "${inputs.params.email}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-create-cluster-kaniko-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-create-cluster-kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-apply
    image: "${inputs.params.kfctl_image}"
    imagePullPolicy: IfNotPresent
    workingDir: "${inputs.params.pvc_mount_path}/${inputs.params.app_dir}"
#    command: ["/bin/sleep", "infinity"]
    command: ["/usr/local/bin/kfctl"]
    args:
    - "apply"
    - "${inputs.params.platform}"
    - "--verbose"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-create-cluster-gcp-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-create-cluster-gcp-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-configure-kubectl
    image: "${inputs.params.kfctl_image}"
    imagePullPolicy: IfNotPresent
    workingDir: "${inputs.params.pvc_mount_path}"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "--project"
    - "${inputs.params.project}"
    - "container"
    - "clusters"
    - "--zone"
    - "${inputs.params.zone}"
    - "get-credentials"
    - "${inputs.params.cluster}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-create-cluster-gcp-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: ci-create-cluster-client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-create-cluster-gcp-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  volumes:
  - name: ci-create-cluster-kaniko-secret
    secret:
      secretName: ci-create-cluster-kaniko-secret
  - name: ci-create-cluster-docker-secret
    secret:
      secretName: ci-create-cluster-docker-secret
  - name: ci-create-cluster-gcp-secret
    secret:
      secretName: ci-create-cluster-gcp-secret
  - name: kubeflow
    persistentVolumeClaim:
      claimName: ci-create-cluster-persistent-volume-claim
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/params.yaml", `
varReference:
- path: metadata/name
  kind: TaskRun
- path: spec/inputs/params/value
  kind: TaskRun
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/create-cluster-task/params.env", `
namespace=kubeflow-ci
project=constant-cubist-173123
app_dir=/kubeflow/kubeflow-ci
zone=us-west1-a
email=foo@bar.com
configPath=https://raw.githubusercontent.com/kubeflow/kubeflow/master/bootstrap/config/ci-cluster.yaml
platform=all
cluster=kubeflow-ci
pvc_mount_path=/kubeflow
kfctl_image=gcr.io/constant-cubist-173123/kfctl@sha256:ab0c4986322e3e6a755056278c7270983b0f3bdc0751aefff075fb2b3d0c3254
`)
	th.writeK("/manifests/ci/ci-pipeline/overlays/create-cluster-task", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
- task.yaml
namespace: tekton-pipelines
namePrefix: ci-create-cluster-
configMapGenerator:
- name: parameters
  env: params.env
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: project
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: configPath
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.configPath
- name: app_dir
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.app_dir
- name: zone
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.zone
- name: email
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.email
- name: platform
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.platform
- name: cluster
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cluster
- name: kfctl_image
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.kfctl_image
- name: pvc_mount_path
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pvc_mount_path
configurations:
- params.yaml
`)
	th.writeF("/manifests/ci/ci-pipeline/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: ci-pipeline
spec:
  params: []
  resources: []
  tasks: []
`)
	th.writeK("/manifests/ci/ci-pipeline/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
namespace: kubeflow-ci
`)
}

func TestCiPipelineOverlaysCreateClusterTask(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline/overlays/create-cluster-task")
	writeCiPipelineOverlaysCreateClusterTask(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline/overlays/create-cluster-task"
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
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
