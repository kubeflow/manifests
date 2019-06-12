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

func writeKatibV1Alpha2SuggestionBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-bayesianoptimization-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-bayesianoptimization-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-hyperband-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: suggestion-hyperband
  name: katib-suggestion-hyperband
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: suggestion-hyperband
      name: katib-suggestion-hyperband
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-suggestion-hyperband
        ports:
        - containerPort: 6789
          name: api
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-hyperband-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: suggestion-hyperband
  name: katib-suggestion-hyperband
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: suggestion-hyperband
  type: ClusterIP
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-grid-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-grid-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-nasrl-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-nasrl-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-random-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-random-service.yaml", `
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
	th.writeK("/manifests/katib-v1alpha2/suggestion/base", `
namespace: kubeflow
resources:
- suggestion-bayesianoptimization-deployment.yaml
- suggestion-bayesianoptimization-service.yaml
- suggestion-hyperband-deployment.yaml
- suggestion-hyperband-service.yaml
- suggestion-grid-deployment.yaml
- suggestion-grid-service.yaml
- suggestion-nasrl-deployment.yaml
- suggestion-nasrl-service.yaml
- suggestion-random-deployment.yaml
- suggestion-random-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random
    newTag: v0.1.2-alpha-280-gb0e0dd5
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl
    newTag: v0.1.2-alpha-280-gb0e0dd5
`)
}

func TestKatibV1Alpha2SuggestionBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/suggestion/base")
	writeKatibV1Alpha2SuggestionBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/suggestion/base"
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
