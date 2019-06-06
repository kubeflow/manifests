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

	th.writeF("/manifests/katib/v1alpha2/base/katib-db-pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: katib-mysql
  namespace: kubeflow
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-ui-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: ui
  name: katib-ui
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: ui
      name: katib-ui
    spec:
      containers:
      - command:
        - ./katib-ui
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-ui
        ports:
        - containerPort: 80
          name: ui
      serviceAccountName: katib-ui
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-ui-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: katib-ui
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - trials
  verbs:
  - '*'
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-ui
  namespace: kubeflow
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: katib-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-ui
subjects:
- kind: ServiceAccount
  name: katib-ui
  namespace: kubeflow
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: ui
  name: katib-ui
  namespace: kubeflow
spec:
  ports:
  - name: ui
    port: 80
    protocol: TCP
  selector:
    app: katib
    component: ui
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/metrics-collector-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
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
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: metrics-collector
  namespace: kubeflow
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: metrics-collector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metrics-collector
subjects:
- kind: ServiceAccount
  name: metrics-collector
  namespace: kubeflow
`)
	th.writeF("/manifests/katib/v1alpha2/base/metrics-collector-template-configmap.yaml", `
apiVersion: v1
data:
  defaultMetricsCollectorTemplate.yaml: |-
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
                image: gcr.io/kubeflow-images-public/katib/v1alpha2/metrics-collector:v0.1.2-alpha-280-gb0e0dd5
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
kind: ConfigMap
metadata:
  name: metrics-collector-template
  namespace: kubeflow
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-controller-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: katib-controller
  name: katib-controller
  namespace: kubeflow
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
      containers:
      - command:
        - ./katib-controller
        env:
        - name: KATIB_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-controller:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-controller
        ports:
        - containerPort: 443
          name: webhook
          protocol: TCP
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      serviceAccountName: katib-controller
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: katib-controller
`)
	th.writeF("/manifests/katib/v1alpha2/base/experiment-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: experiments.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    categories:
    - all
    - kubeflow
    - katib
    kind: Experiment
    plural: experiments
    singular: experiment
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha2
`)
	th.writeF("/manifests/katib/v1alpha2/base/trial-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: trials.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    categories:
    - all
    - kubeflow
    - katib
    kind: Trial
    plural: trials
    singular: trial
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha2
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-controller-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
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
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - '*'
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
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  verbs:
  - '*'
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-controller
  namespace: kubeflow
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: katib-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-controller
subjects:
- kind: ServiceAccount
  name: katib-controller
  namespace: kubeflow
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-controller-secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: katib-controller
  namespace: kubeflow
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-controller-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-controller
  namespace: kubeflow
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    app: katib-controller
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-bayesianoptimization-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: suggestion-bayesianoptimization
  name: katib-suggestion-bayesianoptimization
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: suggestion-bayesianoptimization
      name: katib-suggestion-bayesianoptimization
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-suggestion-bayesianoptimization
        ports:
        - containerPort: 6789
          name: api
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-bayesianoptimization-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: suggestion-bayesianoptimization
  name: katib-suggestion-bayesianoptimization
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: suggestion-bayesianoptimization
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-grid-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: suggestion-grid
  name: katib-suggestion-grid
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: suggestion-grid
      name: katib-suggestion-grid
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-suggestion-grid
        ports:
        - containerPort: 6789
          name: api
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-grid-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: suggestion-grid
  name: katib-suggestion-grid
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: suggestion-grid
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-nasrl-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: suggestion-nasrl
  name: katib-suggestion-nasrl
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: suggestion-nasrl
      name: katib-suggestion-nasrl
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl:v0.1.2-alpha-280-gb0e0dd5
        name: katib-suggestion-nasrl
        ports:
        - containerPort: 6789
          name: api
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-nasrl-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: suggestion-nasrl
  name: katib-suggestion-nasrl
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: suggestion-nasrl
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-random-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: suggestion-random
  name: katib-suggestion-random
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: suggestion-random
      name: katib-suggestion-random
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-suggestion-random
        ports:
        - containerPort: 6789
          name: api
`)
	th.writeF("/manifests/katib/v1alpha2/base/suggestion-random-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: suggestion-random
  name: katib-suggestion-random
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: suggestion-random
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-manager-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: manager
  name: katib-manager
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: manager
      name: katib-manager
    spec:
      containers:
      - command:
        - ./katib-manager
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              key: MYSQL_ROOT_PASSWORD
              name: katib-db-secrets
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        livenessProbe:
          exec:
            command:
            - /bin/grpc_health_probe
            - -addr=:6789
          initialDelaySeconds: 10
        name: katib-manager
        ports:
        - containerPort: 6789
          name: api
        readinessProbe:
          exec:
            command:
            - /bin/grpc_health_probe
            - -addr=:6789
          initialDelaySeconds: 5
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-manager-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: manager
  name: katib-manager
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: manager
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-manager-rest-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: manager-rest
  name: katib-manager-rest
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: manager-rest
      name: katib-manager-rest
    spec:
      containers:
      - command:
        - ./katib-manager-rest
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-manager-rest
        ports:
        - containerPort: 80
          name: api
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-manager-rest-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: manager-rest
  name: katib-manager-rest
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 80
    protocol: TCP
  selector:
    app: katib
    component: manager-rest
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-db-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: db
  name: katib-db
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: db
      name: katib-db
    spec:
      containers:
      - args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              key: MYSQL_ROOT_PASSWORD
              name: katib-db-secrets
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "true"
        - name: MYSQL_DATABASE
          value: katib
        image: mysql:8.0.3
        name: katib-db
        ports:
        - containerPort: 3306
          name: dbapi
        readinessProbe:
          exec:
            command:
            - /bin/bash
            - -c
            - mysql -D $$MYSQL_DATABASE -p$$MYSQL_ROOT_PASSWORD -e 'SELECT 1'
          initialDelaySeconds: 5
          periodSeconds: 2
          timeoutSeconds: 1
        volumeMounts:
        - mountPath: /var/lib/mysql
          name: katib-mysql
      volumes:
      - name: katib-mysql
        persistentVolumeClaim:
          claimName: katib-mysql
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: katib-db-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
	th.writeF("/manifests/katib/v1alpha2/base/katib-db-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: db
  name: katib-db
  namespace: kubeflow
spec:
  ports:
  - name: dbapi
    port: 3306
    protocol: TCP
  selector:
    app: katib
    component: db
  type: ClusterIP
`)
	th.writeF("/manifests/katib/v1alpha2/base/trial-template.yaml", `
apiVersion: v1
data:
  defaultTrialTemplate.yaml: |-
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
kind: ConfigMap
metadata:
  name: trial-template
  namespace: kubeflow
`)
	th.writeF("/manifests/katib/v1alpha2/base/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeF("/manifests/katib/v1alpha2/base/params.env", `
clusterDomain=cluster.local
`)
	th.writeK("/manifests/katib/v1alpha2/base", `
namespace: kubeflow
resources:
- experiment-crd.yaml
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
- suggestion-nasrl-deployment.yaml
- suggestion-nasrl-service.yaml
- suggestion-random-deployment.yaml
- suggestion-random-service.yaml
- trial-crd.yaml
- trial-template.yaml
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-controller
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/metrics-collector
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: mysql
    newTag: 8.0.3
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

func TestKatibV1Alpha2Base(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib/v1alpha2/base")
	writeKatibV1Alpha2Base(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib/v1alpha2/base"
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
