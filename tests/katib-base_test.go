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

func writeKatibBase(th *KustTestHarness) {
  th.writeF("/manifests/katib/base/katib-db-pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: katib-mysql
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
  th.writeF("/manifests/katib/base/katib-ui-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-ui
  labels:
    component: ui
spec:
  replicas: 1
  template:
    metadata:
      name: katib-ui
      labels:
        component: ui
    spec:
      containers:
      - name: katib-ui
        image: gcr.io/kubeflow-images-public/katib/katib-ui:v0.1.2-alpha-156-g4ab3dbd
        command:
          - './katib-ui'
        ports:
        - name: ui
          containerPort: 80
      serviceAccountName: katib-ui
`)
  th.writeF("/manifests/katib/base/katib-ui-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-ui
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - "*"
- apiGroups:
  - kubeflow.org
  resources:
  - studyjobs
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-ui
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-ui
subjects:
- kind: ServiceAccount
  name: katib-ui
`)
  th.writeF("/manifests/katib/base/katib-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-ui
  labels:
    component: ui
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: ui
  selector:
    component: ui
`)
  th.writeF("/manifests/katib/base/katib-ui-virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: katib-ui
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /katib/
    rewrite:
      uri: /katib/
    route:
    - destination:
        host: katib-ui.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
  th.writeF("/manifests/katib/base/metrics-collector-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: metrics-collector
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - "*"
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: metrics-collector
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: metrics-collector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metrics-collector
subjects:
- kind: ServiceAccount
  name: metrics-collector
`)
  th.writeF("/manifests/katib/base/metrics-collector-template-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: metricscollector-template
data:
  defaultMetricsCollectorTemplate.yaml : |-
    apiVersion: batch/v1beta1
    kind: CronJob
    metadata:
      name: {{.WorkerID}}
      namespace: {{.NameSpace}}
    spec:
      schedule: "*/1 * * * *"
      successfulJobsHistoryLimit: 0
      failedJobsHistoryLimit: 1
      jobTemplate:
        spec:
          backoffLimit: 0
          template:
            spec:
              serviceAccountName: metrics-collector
              containers:
              - name: {{.WorkerID}}
                image: gcr.io/kubeflow-images-public/katib/metrics-collector:v0.1.2-alpha-156-g4ab3dbd
                args:
                - "./metricscollector"
                - "-s"
                - "{{.StudyID}}"
                - "-t"
                - "{{.TrialID}}"
                - "-w"
                - "{{.WorkerID}}"
                - "-k"
                - "{{.WorkerKind}}"
                - "-n"
                - "{{.NameSpace}}"
                - "-m"
                - "{{.ManagerSerivce}}"
              restartPolicy: Never
`)
  th.writeF("/manifests/katib/base/studyjob-controller-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: studyjob-controller
  labels:
    app: studyjob-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: studyjob-controller
  template:
    metadata:
      labels:
        app: studyjob-controller
    spec:
      serviceAccountName: studyjob-controller
      containers:
      - name: studyjob-controller
        image: gcr.io/kubeflow-images-public/katib/studyjob-controller:v0.1.2-alpha-156-g4ab3dbd
        imagePullPolicy: Always
        ports:
        - containerPort: 443
          name: validating
          protocol: TCP
        env:
        - name: VIZIER_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
`)
  th.writeF("/manifests/katib/base/studyjob-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: studyjobs.kubeflow.org
spec:
  group: kubeflow.org
  version: v1alpha1
  scope: Namespaced
  names:
    kind: StudyJob
    singular: studyjob
    plural: studyjobs
  additionalPrinterColumns:
  - JSONPath: .status.condition
    name: Condition
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
`)
  th.writeF("/manifests/katib/base/studyjob-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: studyjob-controller
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - serviceaccounts
  - services
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - "*"
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - "*"
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - studyjobs
  verbs:
  - "*"
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: studyjob-controller
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: studyjob-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: studyjob-controller
subjects:
- kind: ServiceAccount
  name: studyjob-controller
`)
  th.writeF("/manifests/katib/base/studyjob-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: studyjob-controller
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    app: studyjob-controller
`)
  th.writeF("/manifests/katib/base/suggestion-bayesianoptimization-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-suggestion-bayesianoptimization
  labels:
    component: suggestion-bayesianoptimization
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-suggestion-bayesianoptimization
      labels:
        component: suggestion-bayesianoptimization
    spec:
      containers:
      - name: vizier-suggestion-bayesianoptimization
        image: gcr.io/kubeflow-images-public/katib/suggestion-bayesianoptimization:v0.1.2-alpha-156-g4ab3dbd
        ports:
        - name: api
          containerPort: 6789
`)
  th.writeF("/manifests/katib/base/suggestion-bayesianoptimization-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-suggestion-bayesianoptimization
  labels:
    component: suggestion-bayesianoptimization
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    component: suggestion-bayesianoptimization
`)
  th.writeF("/manifests/katib/base/suggestion-grid-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-suggestion-grid
  labels:
    component: suggestion-grid
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-suggestion-grid
      labels:
        component: suggestion-grid
    spec:
      containers:
      - name: vizier-suggestion-grid
        image: gcr.io/kubeflow-images-public/katib/suggestion-grid:v0.1.2-alpha-156-g4ab3dbd
        ports:
        - name: api
          containerPort: 6789
`)
  th.writeF("/manifests/katib/base/suggestion-grid-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-suggestion-grid
  labels:
    component: suggestion-grid
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    component: suggestion-grid
`)
  th.writeF("/manifests/katib/base/suggestion-hyperband-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-suggestion-hyperband
  labels:
    component: suggestion-hyperband
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-suggestion-hyperband
      labels:
        component: suggestion-hyperband
    spec:
      containers:
      - name: vizier-suggestion-hyperband
        image: gcr.io/kubeflow-images-public/katib/suggestion-hyperband:v0.1.2-alpha-156-g4ab3dbd
        ports:
        - name: api
          containerPort: 6789
`)
  th.writeF("/manifests/katib/base/suggestion-hyperband-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-suggestion-hyperband
  labels:
    component: suggestion-hyperband
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    component: suggestion-hyperband
`)
  th.writeF("/manifests/katib/base/suggestion-nasrl-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-suggestion-nasrl
  labels:
    component: suggestion-nasrl
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-suggestion-nasrl
      labels:
        component: suggestion-nasrl
    spec:
      containers:
      - name: vizier-suggestion-nasrl
        image: gcr.io/kubeflow-images-public/katib/suggestion-nasrl:v0.1.2-alpha-156-g4ab3dbd
        ports:
        - name: api
          containerPort: 6789
`)
  th.writeF("/manifests/katib/base/suggestion-nasrl-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-suggestion-nasrl
  labels:
    component: suggestion-nasrl
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    component: suggestion-nasrl
`)
  th.writeF("/manifests/katib/base/suggestion-random-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-suggestion-random
  labels:
    component: suggestion-random
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-suggestion-random
      labels:
        component: suggestion-random
    spec:
      containers:
      - name: vizier-suggestion-random
        image: gcr.io/kubeflow-images-public/katib/suggestion-random:v0.1.2-alpha-156-g4ab3dbd
        ports:
        - name: api
          containerPort: 6789
`)
  th.writeF("/manifests/katib/base/suggestion-random-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-suggestion-random
  labels:
    component: suggestion-random
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    component: suggestion-random
`)
  th.writeF("/manifests/katib/base/vizier-core-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-core
  labels:
    component: core
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-core
      labels:
        component: core
    spec:
      serviceAccountName: vizier-core
      containers:
      - name: vizier-core
        image: gcr.io/kubeflow-images-public/katib/vizier-core:v0.1.2-alpha-156-g4ab3dbd
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: vizier-db-secrets
                key: MYSQL_ROOT_PASSWORD
        command:
          - './vizier-manager'
        ports:
        - name: api
          containerPort: 6789
        readinessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:6789"]
          initialDelaySeconds: 5
        livenessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:6789"]
          initialDelaySeconds: 10
`)
  th.writeF("/manifests/katib/base/vizier-core-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: vizier-core
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: vizier-core
subjects:
- kind: ServiceAccount
  name: vizier-core
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: vizier-core
rules:
  - apiGroups: [""]
    resources: ["pods", "nodes", "nodes/*", "pods/log", "pods/status", "services", "persistentvolumes", "persistentvolumes/status","persistentvolumeclaims","persistentvolumeclaims/status"]
    verbs: ["*"]
  - apiGroups: ["batch"]
    resources: ["jobs", "jobs/status"]
    verbs: ["*"]
  - apiGroups: ["extensions"]
    verbs: ["*"]
    resources: ["ingresses","ingresses/status","deployments","deployments/status"]
  - apiGroups: [""]
    verbs: ["*"]
    resources: ["services"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vizier-core
`)
  th.writeF("/manifests/katib/base/vizier-core-rest-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-core-rest
  labels:
    component: core-rest
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-core-rest
      labels:
        component: core-rest
    spec:
      containers:
      - name: vizier-core-rest
        image: gcr.io/kubeflow-images-public/katib/vizier-core-rest:v0.1.2-alpha-156-g4ab3dbd
        command:
          - './vizier-manager-rest'
        ports:
        - name: api
          containerPort: 80
`)
  th.writeF("/manifests/katib/base/vizier-core-rest-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-core-rest
  labels:
    component: core-rest
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: api
  selector:
    component: core-rest
`)
  th.writeF("/manifests/katib/base/vizier-core-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-core
  labels:
    component: core
spec:
  type: NodePort
  ports:
    - port: 6789
      protocol: TCP
      nodePort: 30678
      name: api
  selector:
    component: core
`)
  th.writeF("/manifests/katib/base/vizier-db-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-db
  labels:
    component: db
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-db
      labels:
        component: db
    spec:
      containers:
      - name: vizier-db
        image: mysql:8.0.3
        args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: vizier-db-secrets
                key: MYSQL_ROOT_PASSWORD
          - name: MYSQL_ALLOW_EMPTY_PASSWORD
            value: "true"
          - name: MYSQL_DATABASE
            value: "vizier"
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
        - name: katib-mysql
          mountPath: /var/lib/mysql
      volumes:
      - name: katib-mysql
        persistentVolumeClaim:
          claimName: katib-mysql
`)
  th.writeF("/manifests/katib/base/vizier-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: vizier-db-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
  th.writeF("/manifests/katib/base/vizier-db-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-db
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
  th.writeF("/manifests/katib/base/worker-template.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: worker-template
data:
  defaultWorkerTemplate.yaml : |-
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: {{.WorkerID}}
      namespace: {{.NameSpace}}
    spec:
      template:
        spec:
          containers:
          - name: {{.WorkerID}}
            image: alpine
          restartPolicy: Never
`)
  th.writeF("/manifests/katib/base/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
  th.writeF("/manifests/katib/base/params.env", `
clusterDomain=cluster.local
`)
  th.writeK("/manifests/katib/base", `
namespace: kubeflow
resources:
- katib-db-pvc.yaml
- katib-ui-deployment.yaml
- katib-ui-rbac.yaml
- katib-ui-service.yaml
- katib-ui-virtual-service.yaml
- metrics-collector-rbac.yaml
- metrics-collector-template-configmap.yaml
- studyjob-controller-deployment.yaml
- studyjob-crd.yaml
- studyjob-rbac.yaml
- studyjob-service.yaml
- suggestion-bayesianoptimization-deployment.yaml
- suggestion-bayesianoptimization-service.yaml
- suggestion-grid-deployment.yaml
- suggestion-grid-service.yaml
- suggestion-hyperband-deployment.yaml
- suggestion-hyperband-service.yaml
- suggestion-nasrl-deployment.yaml
- suggestion-nasrl-service.yaml
- suggestion-random-deployment.yaml
- suggestion-random-service.yaml
- vizier-core-deployment.yaml
- vizier-core-rbac.yaml
- vizier-core-rest-deployment.yaml
- vizier-core-rest-service.yaml
- vizier-core-service.yaml
- vizier-db-deployment.yaml
- vizier-db-secret.yaml
- vizier-db-service.yaml
- worker-template.yaml
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/vizier-core
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-hyperband
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/katib-ui
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: mysql
    newTag: 8.0.3
  - name: gcr.io/kubeflow-images-public/katib/suggestion-bayesianoptimization
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-grid
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/vizier-core-rest
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/metrics-collector
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/studyjob-controller
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-random
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-nasrl
    newTag: v0.1.2-alpha-157-g3d4cd04
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: namespace
  objref:
    kind: Service
    name: katib-ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
configurations:
- params.yaml
`)
}

func TestKatibBase(t *testing.T) {
  th := NewKustTestHarness(t, "/manifests/katib/base")
  writeKatibBase(th)
  m, err := th.makeKustTarget().MakeCustomizedResMap()
  if err != nil {
    t.Fatalf("Err: %v", err)
  }
  targetPath := "../katib/base"
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
