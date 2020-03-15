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

func writeJupyterJupyterWebAppBase_v3(th *KustTestHarness) {
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-role
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - create
  - delete
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/finalizers
  - poddefaults
  verbs:
  - get
  - list
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - create
  - delete
  - get
  - list
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - list
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-notebook-ui-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-notebook-ui-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/finalizers
  - poddefaults
  verbs:
  - get
  - list
  - create
  - delete

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-notebook-ui-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/finalizers
  - poddefaults
  verbs:
  - get
  - list
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/jupyter-web-app        
        name: jupyter-web-app
        ports:
        - containerPort: 5000
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      serviceAccountName: service-account
      volumes:
      - configMap:
          name: jupyter-web-app-config
        name: config-volume
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: jupyter-notebook-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: jupyter-notebook-role
subjects:
- kind: ServiceAccount
  name: jupyter-notebook
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: jupyter-notebook-role
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - secrets
  - services
  verbs:
  - '*'
- apiGroups:
  - ""
  - apps
  - extensions
  resources:
  - deployments
  - replicasets
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: webapp_mapping
      prefix: /$(prefix)/
      service: jupyter-web-app-service.$(namespace)
      add_request_headers:
        x-forwarded-prefix: /jupyter
  labels:
    run: jupyter-web-app
  name: service
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 5000
  type: ClusterIP
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/deployment_patch.yaml", `
# TODO(https://github.com/kubeflow/manifests/issues/774): This is a patch
# that pulls out from core the parts that should be in pulled into stacks.
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: jupyter-web-app        
        env:
        - name: USERID_HEADER
          valueFrom:
            configMapKeyRef:
              name: kubeflow-config
              key: userid-header
        - name: USERID_PREFIX
          valueFrom:
            configMapKeyRef:
              name: kubeflow-config
              key: userid-prefix`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base_v3/configs/spawner_ui_config.yaml", `
# Configuration file for the Jupyter UI.
#
# Each Jupyter UI option is configured by two keys: 'value' and 'readOnly'
# - The 'value' key contains the default value
# - The 'readOnly' key determines if the option will be available to users
#
# If the 'readOnly' key is present and set to 'true', the respective option
# will be disabled for users and only set by the admin. Also when a
# Notebook is POSTED to the API if a necessary field is not present then
# the value from the config will be used.
#
# If the 'readOnly' key is missing (defaults to 'false'), the respective option
# will be available for users to edit.
#
# Note that some values can be templated. Such values are the names of the
# Volumes as well as their StorageClass
spawnerFormDefaults:
  image:
    # The container Image for the user's Jupyter Notebook
    # If readonly, this value must be a member of the list below
    value: gcr.io/kubeflow-images-public/tensorflow-1.14.0-notebook-cpu:v-base-ef41372-1177829795472347138
    # The list of available standard container Images
    options:
      - gcr.io/kubeflow-images-public/tensorflow-1.15.2-notebook-cpu:1.0.0
      - gcr.io/kubeflow-images-public/tensorflow-1.15.2-notebook-gpu:1.0.0
      - gcr.io/kubeflow-images-public/tensorflow-2.1.0-notebook-cpu:1.0.0
      - gcr.io/kubeflow-images-public/tensorflow-2.1.0-notebook-gpu:1.0.0
    # By default, custom container Images are allowed
    # Uncomment the following line to only enable standard container Images
    readOnly: false
  cpu:
    # CPU for user's Notebook
    value: '0.5'
    readOnly: false
  memory:
    # Memory for user's Notebook
    value: 1.0Gi
    readOnly: false
  workspaceVolume:
    # Workspace Volume to be attached to user's Notebook
    # Each Workspace Volume is declared with the following attributes:
    # Type, Name, Size, MountPath and Access Mode
    value:
      type:
        # The Type of the Workspace Volume
        # Supported values: 'New', 'Existing'
        value: New
      name:
        # The Name of the Workspace Volume
        # Note that this is a templated value. Special values:
        # {notebook-name}: Replaced with the name of the Notebook. The frontend
        #                  will replace this value as the user types the name
        value: 'workspace-{notebook-name}'
      size:
        # The Size of the Workspace Volume (in Gi)
        value: '10Gi'
      mountPath:
        # The Path that the Workspace Volume will be mounted
        value: /home/jovyan
      accessModes:
        # The Access Mode of the Workspace Volume
        # Supported values: 'ReadWriteOnce', 'ReadWriteMany', 'ReadOnlyMany'
        value: ReadWriteOnce
      class:
        # The StrageClass the PVC will use if type is New. Special values are:
        # {none}: default StorageClass
        # {empty}: empty string ""
        value: '{none}'
    readOnly: false
  dataVolumes:
    # List of additional Data Volumes to be attached to the user's Notebook
    value: []
    # Each Data Volume is declared with the following attributes:
    # Type, Name, Size, MountPath and Access Mode
    #
    # For example, a list with 2 Data Volumes:
    # value:
    #   - value:
    #       type:
    #         value: New
    #       name:
    #         value: '{notebook-name}-vol-1'
    #       size:
    #         value: '10Gi'
    #       class:
    #         value: standard
    #       mountPath:
    #         value: /home/jovyan/vol-1
    #       accessModes:
    #         value: ReadWriteOnce
    #       class:
    #         value: {none}
    #   - value:
    #       type:
    #         value: New
    #       name:
    #         value: '{notebook-name}-vol-2'
    #       size:
    #         value: '10Gi'
    #       mountPath:
    #         value: /home/jovyan/vol-2
    #       accessModes:
    #         value: ReadWriteMany
    #       class:
    #         value: {none}
    readOnly: false
  gpus:
    # Number of GPUs to be assigned to the Notebook Container
    value:
      # values: "none", "1", "2", "4", "8"
      num: "none"
      # Determines what the UI will show and send to the backend
      vendors:
        - limitsKey: "nvidia.com/gpu"
          uiName: "NVIDIA"
      # Values: "" or a `+"`"+`limits-key`+"`"+` from the vendors list
      vendor: ""
    readOnly: false
  shm:
    value: true
    readOnly: false
  configurations:
    # List of labels to be selected, these are the labels from PodDefaults
    # value:
    #   - add-gcp-secret
    #   - default-editor
    value: []
    readOnly: false`)
	th.writeK("/manifests/jupyter/jupyter-web-app/base_v3", `
# TODO(https://github.com/kubeflow/manifests/issues/774):
# This is a new kustomization file intended to get rid of the
# need to rely on kfctl to build the kustomization.yaml file.
# We might want to eventually move it to jupyter/jupyter-web-app/kustomization.yaml
# We currently don't do that because we don't want to interfere with existing behavior.
#
# This kustomization.yaml file doesn't depend on base/kustomization.yaml
# because that file contains changes that won't work with the new stack kustomize
# packages that we want to define. For example, we can't define vars namespace, clusterDomain
# etc... because we want those to be defined at the stack level and reused across applications.
# We don't want to modify jupyter-web-app/kustomization.yaml because that would
# break the existing KFDef files. So we want to make the stacks work 
# and then replace it.
apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  app.kubernetes.io/component: jupyter-web-app
  app.kubernetes.io/instance: jupyter-web-app-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/name: jupyter-web-app
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
kind: Kustomization
namePrefix: jupyter-web-app-
namespace: kubeflow
commonLabels:
  app: jupyter-web-app
  kustomize.component: jupyter-web-app
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
- ../overlays/istio
- ../overlays/application
configMapGenerator:
# We need the name to be unique without the suffix because the original name is what
# gets used with patches
- name: jupyter-web-app-config
  files:
  - configs/spawner_ui_config.yaml
patchesStrategicMerge:
- deployment_patch.yaml`)
}

func TestJupyterJupyterWebAppBase_v3(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/jupyter/overlays/base_v3")
	writeJupyterJupyterWebAppBase_v3(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../jupyter/jupyter-web-app/base_v3"
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
