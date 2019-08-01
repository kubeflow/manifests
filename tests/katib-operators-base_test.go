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

func writeKatibOperatorsBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-controller-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-controller
  labels:
    app: katib-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib-controller
  template:
    metadata:
      labels:
        app: katib-controller
    spec:
      serviceAccountName: katib-controller
      containers:
      - name: katib-controller
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-controller:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        command: ["./katib-controller"]
        ports:
        - containerPort: 443
          name: webhook
          protocol: TCP
        env:
        - name: KATIB_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: katib-controller
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-controller-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-controller
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - serviceaccounts
  - services
  - secrets
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
  - mutatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - experiments/status
  - trials
  - trials/status
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
  name: katib-controller
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-controller
subjects:
- kind: ServiceAccount
  name: katib-controller
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-controller-secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: katib-controller
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-controller-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-controller
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    app: katib-controller
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-db-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-db
  labels:
    app: katib
    component: db
spec:
  replicas: 1
  template:
    metadata:
      name: katib-db
      labels:
        app: katib
        component: db
    spec:
      containers:
      - name: katib-db
        image: mysql:8.0.3
        args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: katib-db-secrets
                key: MYSQL_ROOT_PASSWORD
          - name: MYSQL_ALLOW_EMPTY_PASSWORD
            value: "true"
          - name: MYSQL_DATABASE
            value: "katib"
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
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-db-pvc.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: katib-db-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-db-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-db
  labels:
    app: katib
    component: db
spec:
  type: ClusterIP
  ports:
    - port: 3306
      protocol: TCP
      name: dbapi
  selector:
    app: katib
    component: db
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-manager-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-manager
  labels:
    app: katib
    component: manager
spec:
  replicas: 1
  template:
    metadata:
      name: katib-manager
      labels:
        app: katib
        component: manager
    spec:
      containers:
      - name: katib-manager
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: katib-db-secrets
                key: MYSQL_ROOT_PASSWORD
        command:
          - './katib-manager'
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
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-manager-rest-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-manager-rest
  labels:
    app: katib
    component: manager-rest
spec:
  replicas: 1
  template:
    metadata:
      name: katib-manager-rest
      labels:
        app: katib
        component: manager-rest
    spec:
      containers:
      - name: katib-manager-rest
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        command:
          - './katib-manager-rest'
        ports:
        - name: api
          containerPort: 80
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-manager-rest-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-manager-rest
  labels:
    app: katib
    component: manager-rest
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: api
  selector:
    app: katib
    component: manager-rest
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-manager-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-manager
  labels:
    app: katib
    component: manager
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: manager
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-ui-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  replicas: 1
  template:
    metadata:
      name: katib-ui
      labels:
        app: katib
        component: ui
    spec:
      containers:
      - name: katib-ui
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        command:
          - './katib-ui'
        ports:
        - name: ui
          containerPort: 80
      serviceAccountName: katib-ui
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-ui-rbac.yaml", `
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
  - experiments
  - trials
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
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/katib-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: ui
  selector:
    app: katib
    component: ui
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/metrics-collector-rbac.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/metrics-collector-template-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: metrics-collector-template
data:
  defaultMetricsCollectorTemplate.yaml : |-
    apiVersion: batch/v1beta1
    kind: CronJob
    metadata:
      name: {{.Trial}}
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
              - name: {{.Trial}}
                image: gcr.io/kubeflow-images-public/katib/v1alpha2/metrics-collector:v0.1.2-alpha-289-g14dad8b
                imagePullPolicy: IfNotPresent
                command: ["./metricscollector"]
                args:
                - "-e"
                - "{{.Experiment}}"
                - "-t"
                - "{{.Trial}}"
                - "-k"
                - "{{.JobKind}}"
                - "-n"
                - "{{.NameSpace}}"
                - "-m"
                - "{{.ManagerService}}"
                - "-mn"
                - "{{.MetricNames}}"
              restartPolicy: Never
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-bayesianoptimization-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-suggestion-bayesianoptimization
  labels:
    app: katib
    component: suggestion-bayesianoptimization
spec:
  replicas: 1
  template:
    metadata:
      name: katib-suggestion-bayesianoptimization
      labels:
        app: katib
        component: suggestion-bayesianoptimization
    spec:
      containers:
      - name: katib-suggestion-bayesianoptimization
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-bayesianoptimization-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-bayesianoptimization
  labels:
    app: katib
    component: suggestion-bayesianoptimization
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-bayesianoptimization
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-grid-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-suggestion-grid
  labels:
    app: katib
    component: suggestion-grid
spec:
  replicas: 1
  template:
    metadata:
      name: katib-suggestion-grid
      labels:
        app: katib
        component: suggestion-grid
    spec:
      containers:
      - name: katib-suggestion-grid
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-grid-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-grid
  labels:
    app: katib
    component: suggestion-grid
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-grid
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-hyperband-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-suggestion-hyperband
  labels:
    app: katib
    component: suggestion-hyperband
spec:
  replicas: 1
  template:
    metadata:
      name: katib-suggestion-hyperband
      labels:
        app: katib
        component: suggestion-hyperband
    spec:
      containers:
      - name: katib-suggestion-hyperband
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-hyperband-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-hyperband
  labels:
    app: katib
    component: suggestion-hyperband
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-hyperband
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-nasrl-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-suggestion-nasrl
  labels:
    app: katib
    component: suggestion-nasrl
spec:
  replicas: 1
  template:
    metadata:
      name: katib-suggestion-nasrl
      labels:
        app: katib
        component: suggestion-nasrl
    spec:
      containers:
      - name: katib-suggestion-nasrl
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl:v0.1.2-alpha-289-g14dad8b
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-nasrl-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-nasrl
  labels:
    app: katib
    component: suggestion-nasrl
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-nasrl
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-random-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-suggestion-random
  labels:
    app: katib
    component: suggestion-random
spec:
  replicas: 1
  template:
    metadata:
      name: katib-suggestion-random
      labels:
        app: katib
        component: suggestion-random
    spec:
      containers:
      - name: katib-suggestion-random
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/suggestion-random-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-random
  labels:
    app: katib
    component: suggestion-random
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-random
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/trial-template.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: trial-template
data:
  defaultTrialTemplate.yaml : |-
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: {{.Trial}}
      namespace: {{.NameSpace}}
    spec:
      template:
        spec:
          containers:
          - name: {{.Trial}}
            image: alpine
          restartPolicy: Never
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/params.yaml", `
varReference:
- path: data/config
  kind: ConfigMap
- path: data/config
  kind: Deployment
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/katib-v1alpha2/katib-operators/base/params.env", `
clusterDomain=cluster.local
`)
	th.writeK("/manifests/katib-v1alpha2/katib-operators/base", `
namespace: kubeflow
resources:
- katib-controller-deployment.yaml
- katib-controller-rbac.yaml
- katib-controller-secret.yaml
- katib-controller-service.yaml
- katib-db-deployment.yaml
- katib-db-pvc.yaml
- katib-db-secret.yaml
- katib-db-service.yaml
- katib-manager-deployment.yaml
- katib-manager-rest-deployment.yaml
- katib-manager-rest-service.yaml
- katib-manager-service.yaml
- katib-ui-deployment.yaml
- katib-ui-rbac.yaml
- katib-ui-service.yaml
- metrics-collector-rbac.yaml
- metrics-collector-template-configmap.yaml
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
- trial-template.yaml
configMapGenerator:
- name: katib-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-controller
    newTag: v0.6.0-rc.0
  - name: mysql
    newTag: 8.0.3
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/metrics-collector
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl
    newTag: v0.6.0-rc.0
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: katib-parameters
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

func TestKatibOperatorsBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-operators/base")
	writeKatibOperatorsBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-operators/base"
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
