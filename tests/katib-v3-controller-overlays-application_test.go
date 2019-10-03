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

func writeKatibV3ControllerApplication(th *KustTestHarness) {
	th.writeF("/manifests/katib/katib-controller/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: $(generateName)
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: Secret
  - group: core
    kind: ServiceAccount
  - group: kubeflow.org
    kind: Experiment
  - group: kubeflow.org
    kind: Suggestion
  - group: kubeflow.org
    kind: Trial
  descriptor:
    type: "katib"
    version: "v1alpha3"
    description: "Katib is a service for hyperparameter tuning and neural architecture search."
    maintainers:
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    owners:
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    keywords:
    - katib
    - katib-controller
    - hyperparameter tuning
    links:
    - description: About
      url: "https://github.com/kubeflow/katib"
  addOwnerRef: true
`)
	th.writeK("/manifests/katib/katib-controller/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: katib-controller-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: katib-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: katib
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: katib
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.7
`)
	th.writeF("/manifests/katib/katib-controller/overlays/application/params.env", `
generateName=
`)
	th.writeF("/manifests/katib/katib-controller/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: katib-config
data:
  metrics-collector-sidecar: |-
    {
      "StdOut": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/file-metrics-collector"
      },
      "File": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/file-metrics-collector"
      },
      "TensorFlowEvent": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/tfevent-metrics-collector"
      }
    }
  suggestion: |-
    {
      "random": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-hyperopt"
      },
      "grid": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-chocolate"
      },
      "hyperband": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-hyperband"
      },
      "bayesianoptimization": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-skopt"
      },
      "tpe": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-hyperopt"
      },
      "nasrl": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-nasrl"
      }
    }
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-deployment.yaml", `
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
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-controller
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
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-rbac.yaml", `
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
  - events
  - namespaces
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
  - apps
  resources:
  - deployments
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
  - suggestions
  - suggestions/status
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
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: katib-controller
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-service.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/katib-db-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-db
  labels:
    app: katib
    component: db
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: db
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
	th.writeF("/manifests/katib/katib-controller/base/katib-db-pvc.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/katib-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: katib-db-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-db-service.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/katib-manager-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-manager
  labels:
    app: katib
    component: manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: manager
  template:
    metadata:
      name: katib-manager
      labels:
        app: katib
        component: manager
    spec:
      containers:
      - name: katib-manager
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-manager
        imagePullPolicy: IfNotPresent
        env:
          - name : DB_NAME
            value: "mysql"
          - name: DB_PASSWORD
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
	th.writeF("/manifests/katib/katib-controller/base/katib-manager-rest-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-manager-rest
  labels:
    app: katib
    component: manager-rest
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: manager-rest
  template:
    metadata:
      name: katib-manager-rest
      labels:
        app: katib
        component: manager-rest
    spec:
      containers:
      - name: katib-manager-rest
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-manager-rest
        imagePullPolicy: IfNotPresent
        command:
          - './katib-manager-rest'
        ports:
        - name: api
          containerPort: 80
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-manager-rest-service.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/katib-manager-service.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/katib-ui-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: ui
  template:
    metadata:
      name: katib-ui
      labels:
        app: katib
        component: ui
    spec:
      containers:
      - name: katib-ui
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-ui
        imagePullPolicy: IfNotPresent
        command:
          - './katib-ui'
        env:
          - name: KATIB_CORE_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
        ports:
        - name: ui
          containerPort: 80
      serviceAccountName: katib-ui
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-ui-rbac.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/katib-ui-service.yaml", `
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
	th.writeF("/manifests/katib/katib-controller/base/trial-template-configmap.yaml", `
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
            image: docker.io/katib/mxnet-mnist-example
            command:
            - "python"
            - "/mxnet/example/image-classification/train_mnist.py"
            - "--batch-size=64"
            {{- with .HyperParameters}}
            {{- range .}}
            - "{{.Name}}={{.Value}}"
            {{- end}}
            {{- end}}
          restartPolicy: Never
`)
	th.writeF("/manifests/katib/katib-controller/base/params.yaml", `
varReference:
- path: data/config
  kind: ConfigMap
- path: data/config
  kind: Deployment
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/katib/katib-controller/base/params.env", `
clusterDomain=cluster.local
`)
	th.writeK("/manifests/katib/katib-controller/base", `
namespace: kubeflow
resources:
- katib-configmap.yaml
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
- trial-template-configmap.yaml
configMapGenerator:
- name: katib-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-controller
    newTag: 7ade03b
  - name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-manager
    newTag: 7ade03b
  - name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-manager-rest
    newTag: 7ade03b
  - name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-ui
    newTag: 7ade03b
  - name: mysql
    newTag: 8.0.3
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

func TestKatibV3ControllerApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib/katib-controller/overlays/application")
	writeKatibV3ControllerApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib/katib-controller/overlays/application"
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
