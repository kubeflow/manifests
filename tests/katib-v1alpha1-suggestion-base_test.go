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

func writeKatibV1Alpha1Suggestion(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-bayesianoptimization-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-bayesianoptimization-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-grid-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-grid-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-hyperband-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-hyperband-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-nasrl-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-nasrl-service.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-random-deployment.yaml", `
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
	th.writeF("/manifests/katib-v1alpha1/suggestion/base/suggestion-random-service.yaml", `
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
	th.writeK("/manifests/katib-v1alpha1/suggestion/base", `
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
  - name: gcr.io/kubeflow-images-public/katib/suggestion-hyperband
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-bayesianoptimization
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-grid
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-random
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/suggestion-nasrl
    newTag: v0.1.2-alpha-157-g3d4cd04
`)
}

func TestKatibV1Alpha1Suggestion(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha1/suggestion/base")
	writeKatibV1Alpha1Suggestion(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha1/suggestion/base"
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
