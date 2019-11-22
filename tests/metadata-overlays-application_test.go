package tests_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
)

func writeMetadataOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/metadata/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: metadata
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: metadata
      app.kubernetes.io/instance: metadata-v0.7.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: metadata
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.7.0
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: ConfigMap
  - group: core
    kind: ServiceAccount
  descriptor:
    type: "metadata"
    version: "alpha"
    description: "Tracking and managing metadata of machine learning workflows in Kubeflow."
    maintainers:
    - name: Zhenghui Wang
      email: zhenghui@google.com
    owners:
    - name: Ajay Gopinathan
      email: ajaygopinathan@google.com
    - name: Zhenghui Wang
      email: zhenghui@google.com
    keywords:
    - "metadata"
    links:
    - description: Docs
      url: "https://www.kubeflow.org/docs/components/misc/metadata/"
  addOwnerRef: true
`)
	th.writeK("/manifests/metadata/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: metadata
  app.kubernetes.io/instance: metadata-v0.7.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: metadata
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.7.0
`)
	th.writeF("/manifests/metadata/base/metadata-db-pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
	th.writeF("/manifests/metadata/base/metadata-db-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: db-configmap
data:
  mysql_database: "metadb"
  mysql_port: "3306"
`)
	th.writeF("/manifests/metadata/base/metadata-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: db-secrets
data:
  username: cm9vdA== # "root"
  password: dGVzdA== # "test"
`)
	th.writeF("/manifests/metadata/base/metadata-db-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: db
  labels:
    component: db
spec:
  replicas: 1
  template:
    metadata:
      name: db
      labels:
        component: db
    spec:
      containers:
      - name: db-container
        image: mysql:8.0.3
        args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: metadata-db-secrets
              key: password
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "true"
        - name: MYSQL_DATABASE
          valueFrom:
            configMapKeyRef:
              name: metadata-db-configmap
              key: mysql_database
        ports:
        - name: dbapi
          containerPort: 3306
        readinessProbe:
          exec:
            command:
            - "/bin/bash"
            - "-c"
            - "mysql -D $$MYSQL_DATABASE -p$$MYSQL_ROOT_PASSWORD -e 'SELECT 1'"
          initialDelaySeconds: 5
          periodSeconds: 2
          timeoutSeconds: 1
        volumeMounts:
        - name: metadata-mysql
          mountPath: /var/lib/mysql
      volumes:
      - name: metadata-mysql
        persistentVolumeClaim:
          claimName: metadata-mysql
`)
	th.writeF("/manifests/metadata/base/metadata-db-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: db
  labels:
    component: db
spec:
  type: ClusterIP
  ports:
    - port: 3306
      protocol: TCP
      name: dbapi
  selector:
    component: db
`)
	th.writeF("/manifests/metadata/base/metadata-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
  labels:
    component: server
spec:
  replicas: 1
  selector:
    matchLabels:
      component: server
  template:
    metadata:
      labels:
        component: server
    spec:
      containers:
      - name: container
        image: gcr.io/kubeflow-images-public/metadata:v0.1.11
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: metadata-db-secrets
              key: password
        - name: MYSQL_USER_NAME
          valueFrom:
            secretKeyRef:
              name: metadata-db-secrets
              key: username
        - name: MYSQL_DATABASE
          valueFrom:
            configMapKeyRef:
              name: metadata-db-configmap
              key: mysql_database
        - name: MYSQL_PORT
          valueFrom:
            configMapKeyRef:
              name: metadata-db-configmap
              key: mysql_port
        command: ["./server/server",
                  "--http_port=8080",
                  "--mysql_service_host=metadata-db.kubeflow",
                  "--mysql_service_port=$(MYSQL_PORT)",
                  "--mysql_service_user=$(MYSQL_USER_NAME)",
                  "--mysql_service_password=$(MYSQL_ROOT_PASSWORD)",
                  "--mlmd_db_name=$(MYSQL_DATABASE)"]
        ports:
        - name: backendapi
          containerPort: 8080

        readinessProbe:
          httpGet:
            path: /api/v1alpha1/artifact_types
            port: backendapi
            httpHeaders:
            - name: ContentType
              value: application/json
          initialDelaySeconds: 3
          periodSeconds: 5
          timeoutSeconds: 2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grpc-deployment
  labels:
    component: grpc-server
spec:
  replicas: 1
  selector:
    matchLabels:
      component: grpc-server
  template:
    metadata:
      labels:
        component: grpc-server
    spec:
      containers:
        - name: container
          image: gcr.io/tfx-oss-public/ml_metadata_store_server:0.15.1
          env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: metadata-db-secrets
                key: password
          - name: MYSQL_USER_NAME
            valueFrom:
              secretKeyRef:
                name: metadata-db-secrets
                key: username
          - name: MYSQL_DATABASE
            valueFrom:
              configMapKeyRef:
                name: metadata-db-configmap
                key: mysql_database
          - name: MYSQL_PORT
            valueFrom:
              configMapKeyRef:
                name: metadata-db-configmap
                key: mysql_port
          command: ["/bin/metadata_store_server"]
          args: ["--grpc_port=8080",
                 "--mysql_config_host=metadata-db.kubeflow",
                 "--mysql_config_database=$(MYSQL_DATABASE)",
                 "--mysql_config_port=$(MYSQL_PORT)",
                 "--mysql_config_user=$(MYSQL_USER_NAME)",
                 "--mysql_config_password=$(MYSQL_ROOT_PASSWORD)"
          ]
          ports:
            - name: grpc-backendapi
              containerPort: 8080
`)
	th.writeF("/manifests/metadata/base/metadata-service.yaml", `
kind: Service
apiVersion: v1
metadata:
  labels:
    app: metadata
  name: service
spec:
  selector:
    component: server
  type: ClusterIP
  ports:
  - port: 8080
    protocol: TCP
    name: backendapi
---
kind: Service
apiVersion: v1
metadata:
  labels:
    app: grpc-metadata
  name: grpc-service
spec:
  selector:
    component: grpc-server
  type: ClusterIP
  ports:
    - port: 8080
      protocol: TCP
      name: grpc-backendapi
`)
	th.writeF("/manifests/metadata/base/metadata-ui-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ui
  labels:
    app: metadata-ui
spec:
  selector:
    matchLabels:
      app: metadata-ui
  template:
    metadata:
      name: ui
      labels:
        app: metadata-ui
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/metadata-frontend:v0.1.8
        imagePullPolicy: IfNotPresent
        name: metadata-ui
        ports:
        - containerPort: 3000
      serviceAccountName: ui

`)
	th.writeF("/manifests/metadata/base/metadata-ui-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: metadata-ui
  name: ui
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  verbs:
  - create
  - get
  - list
- apiGroups:
  - "kubeflow.org"
  resources:
  - viewers
  verbs:
  - create
  - get
  - list
  - watch
  - delete
`)
	th.writeF("/manifests/metadata/base/metadata-ui-rolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  labels:
    app: metadata-ui
  name: ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ui
subjects:
- kind: ServiceAccount
  name: ui
  namespace: kubeflow
`)
	th.writeF("/manifests/metadata/base/metadata-ui-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ui
`)
	th.writeF("/manifests/metadata/base/metadata-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: ui
  labels:
    app: metadata-ui
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    app: metadata-ui
`)
	th.writeF("/manifests/metadata/base/metadata-envoy-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: envoy-deployment
  labels:
    component: envoy
spec:
  replicas: 1
  selector:
    matchLabels:
      component: envoy
  template:
    metadata:
      labels:
        component: envoy
    spec:
      containers:
      - name: container
        image: gcr.io/ml-pipeline/envoy:metadata-grpc
        ports:
        - name: md-envoy
          containerPort: 9090
        - name: envoy-admin
          containerPort: 9901
`)
	th.writeF("/manifests/metadata/base/metadata-envoy-service.yaml", `
kind: Service
apiVersion: v1
metadata:
  labels:
    app: metadata
  name: envoy-service
spec:
  selector:
    component: envoy
  type: ClusterIP
  ports:
  - port: 9090
    protocol: TCP
    name: md-envoy
`)
	th.writeF("/manifests/metadata/base/params.env", `
uiClusterDomain=cluster.local
`)
	th.writeK("/manifests/metadata/base", `
namePrefix: metadata-
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  kustomize.component: metadata
configMapGenerator:
- name: ui-parameters
  env: params.env
resources:
- metadata-db-pvc.yaml
- metadata-db-configmap.yaml
- metadata-db-secret.yaml
- metadata-db-deployment.yaml
- metadata-db-service.yaml
- metadata-deployment.yaml
- metadata-service.yaml
- metadata-ui-deployment.yaml
- metadata-ui-role.yaml
- metadata-ui-rolebinding.yaml
- metadata-ui-sa.yaml
- metadata-ui-service.yaml
- metadata-envoy-deployment.yaml
- metadata-envoy-service.yaml
namespace: kubeflow
vars:
- name: ui-namespace
  objref:
    kind: Service
    name: ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: ui-clusterDomain
  objref:
    kind: ConfigMap
    name: ui-parameters
    version: v1
  fieldref:
    fieldpath: data.uiClusterDomain
- name: service
  objref:
    kind: Service
    name: ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: metadata-envoy-service
  objref:
    kind: Service
    name: envoy-service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
images:
- name: gcr.io/kubeflow-images-public/metadata
  newName: gcr.io/kubeflow-images-public/metadata
  newTag: v0.1.11
- name: gcr.io/tfx-oss-public/ml_metadata_store_server
  newName: gcr.io/tfx-oss-public/ml_metadata_store_server
  newTag: 0.15.1
- name: gcr.io/ml-pipeline/envoy
  newName: gcr.io/ml-pipeline/envoy
  newTag: metadata-grpc
- name: mysql
  newName: mysql
  newTag: 8.0.3
- name: gcr.io/kubeflow-images-public/metadata-frontend
  newName: gcr.io/kubeflow-images-public/metadata-frontend
  newTag: v0.1.8
`)
}

func TestMetadataOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/metadata/overlays/application")
	writeMetadataOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../metadata/overlays/application"
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
