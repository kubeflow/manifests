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

func writeAwsAlbIngressControllerOverlaysVpc(th *KustTestHarness) {
	th.writeF("/manifests/aws/aws-alb-ingress-controller/overlays/vpc/vpc.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alb-ingress-controller
spec:
  template:
    spec:
      containers:
        - name: alb-ingress-controller
          args:
            # AWS VPC ID this ingress controller will use to create AWS resources.
            # If unspecified, it will be discovered from ec2metadata.
            - --aws-vpc-id=$(VPC_ID)

            # AWS region this ingress controller will operate in.
            # If unspecified, it will be discovered from ec2metadata.
            # List of regions: http://docs.aws.amazon.com/general/latest/gr/rande.html#vpc_region
            - --aws-region=$(REGION)
`)
	th.writeF("/manifests/aws/aws-alb-ingress-controller/overlays/vpc/params.env", `
vpcId=
region=us-west-2`)
	th.writeK("/manifests/aws/aws-alb-ingress-controller/overlays/vpc", `
bases:
- ../../base
resources:
- vpc.yaml
configMapGenerator:
- name: alb-ingress-controller-parameters
  behavior: merge
  env: params.env
vars:
- name: VPC_ID
  objref:
    kind: ConfigMap
    name: alb-ingress-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.vpcId
- name: REGION
  objref:
    kind: ConfigMap
    name: alb-ingress-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.region
`)
	th.writeF("/manifests/aws/aws-alb-ingress-controller/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: alb-ingress-controller
rules:
  - apiGroups:
      - ""
      - extensions
    resources:
      - configmaps
      - endpoints
      - events
      - ingresses
      - ingresses/status
      - services
    verbs:
      - create
      - get
      - list
      - update
      - watch
      - patch
  - apiGroups:
      - ""
      - extensions
    resources:
      - nodes
      - pods
      - secrets
      - services
      - namespaces
    verbs:
      - get
      - list
      - watch`)
	th.writeF("/manifests/aws/aws-alb-ingress-controller/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: alb-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: alb-ingress-controller
subjects:
  - kind: ServiceAccount
    name: alb-ingress-controller`)
	th.writeF("/manifests/aws/aws-alb-ingress-controller/base/deployment.yaml", `
# Application Load Balancer (ALB) Ingress Controller Deployment Manifest.
# This manifest details sensible defaults for deploying an ALB Ingress Controller.
# GitHub: https://github.com/kubernetes-sigs/aws-alb-ingress-controller
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alb-ingress-controller
  # Namespace the ALB Ingress Controller should run in. Does not impact which
  # namespaces it's able to resolve ingress resource for. For limiting ingress
  # namespace scope, see --watch-namespace.
#  namespace: kubeflow
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: alb-ingress-controller
  template:
    metadata:
      labels:
        app.kubernetes.io/name: alb-ingress-controller
    spec:
      containers:
        - name: alb-ingress-controller
          args:
            # Limit the namespace where this ALB Ingress Controller deployment will
            # resolve ingress resources. If left commented, all namespaces are used.
            # - --watch-namespace=your-k8s-namespace

            # Setting the ingress-class flag below ensures that only ingress resources with the
            # annotation kubernetes.io/ingress.class: "alb" are respected by the controller. You may
            # choose any class you'd like for this controller to respect.
            - --ingress-class=alb

            # REQUIRED
            # Name of your cluster. Used when naming resources created
            # by the ALB Ingress Controller, providing distinction between
            # clusters.
#            - --cluster-name=$(CLUSTER_NAME)
            - --cluster-name=$(CLUSTER_NAME)

            # AWS VPC ID this ingress controller will use to create AWS resources.
            # If unspecified, it will be discovered from ec2metadata.
            # - --aws-vpc-id=vpc-xxxxxx

            # AWS region this ingress controller will operate in.
            # If unspecified, it will be discovered from ec2metadata.
            # List of regions: http://docs.aws.amazon.com/general/latest/gr/rande.html#vpc_region
            # - --aws-region=us-west-1
          # Repository location of the ALB Ingress Controller.
          image: docker.io/amazon/aws-alb-ingress-controller:v1.1.2
          imagePullPolicy: Always
      serviceAccountName: alb-ingress-controller`)
	th.writeF("/manifests/aws/aws-alb-ingress-controller/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: alb-ingress-controller`)
	th.writeF("/manifests/aws/aws-alb-ingress-controller/base/params.env", `
clusterName=`)
	th.writeK("/manifests/aws/aws-alb-ingress-controller/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role.yaml
- cluster-role-binding.yaml
- deployment.yaml
- service-account.yaml
commonLabels:
  kustomize.component: aws-alb-ingress-controller
generatorOptions:
  disableNameSuffixHash: true
images:
- name: docker.io/amazon/aws-alb-ingress-controller
  newName: docker.io/amazon/aws-alb-ingress-controller
  newTag: v1.1.2
configMapGenerator:
- name: alb-ingress-controller-parameters
  env: params.env
vars:
- name: CLUSTER_NAME
  objref:
    kind: ConfigMap
    name: alb-ingress-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterName
`)
}

func TestAwsAlbIngressControllerOverlaysVpc(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/aws-alb-ingress-controller/overlays/vpc")
	writeAwsAlbIngressControllerOverlaysVpc(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/aws-alb-ingress-controller/overlays/vpc"
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
