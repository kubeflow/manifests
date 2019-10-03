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

func writeSuggestionBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-bayesianoptimization-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-bayesianoptimization
  labels:
    app: katib
    component: suggestion-bayesianoptimization
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-bayesianoptimization
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-bayesianoptimization-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-grid-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-grid
  labels:
    app: katib
    component: suggestion-grid
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-grid
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-grid-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-hyperband-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-hyperband
  labels:
    app: katib
    component: suggestion-hyperband
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-hyperband
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-hyperband-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-nasrl-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-nasrl
  labels:
    app: katib
    component: suggestion-nasrl
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-nasrl
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-nasrl-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-random-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-random
  labels:
    app: katib
    component: suggestion-random
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-random
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-random-service.yaml", `
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
	th.writeK("/manifests/katib-v1alpha2/suggestion/base", `
namespace: kubeflow
resources:
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
generatorOptions:
  disableNameSuffixHash: true
images:
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
`)
}

func TestSuggestionBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/suggestion/base")
	writeSuggestionBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/suggestion/base"
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
