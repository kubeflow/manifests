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

func writeKfdefSourceMaster(th *KustTestHarness) {
	th.writeF("/manifests/kfdef/source/master/kfctl_anthos.yaml", `
# This is the config to install Kubeflow on an Anthos.
# The cluster comes with customized Istio installation.

apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-anthos
spec:
  applications:
  # This component is the istio resources for Kubeflow (e.g. gateway), not about installing istio.
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "OFF"
      - name: gatewaySelector
        value: ingress-gke-system
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager        
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: <randomly-generated-id>
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
      - name: admin
        value: johnDoe@acme.com
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
    # An example uri that uses a PR version of the manifest repo.
    # uri: https://github.com/kubeflow/manifests/archive/pull/PR_NUMBER/head.tar.gz
  version: master
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_aws.yaml", `
apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-aws
spec:
  applications:
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/cluster-local-gateway
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "OFF"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/add-anonymous-user-filter
    name: add-anonymous-user-filter
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: webhookNamePrefix
        value: admission-webhook-
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: <randomly-generated-id>
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: mpi-job/mpi-operator
    name: mpi-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: clusterName
        value: kubeflow-aws
      repoRef:
        name: manifests
        path: aws/aws-alb-ingress-controller
    name: aws-alb-ingress-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: aws/nvidia-device-plugin
    name: nvidia-device-plugin
  plugins:
  - kind: KfAwsPlugin
    metadata:
      name: aws
    spec:
      auth:
        basicAuth:
          password:
            name: password
          username: admin
      region: us-west-2
      roles:
      - eksctl-kubeflow-aws-nodegroup-ng-a2-NodeInstanceRole-xxxxxxx
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  version: master
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_aws_cognito.yaml", `
apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-aws-cognito
spec:
  applications:
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/cluster-local-gateway
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "ON"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: webhookNamePrefix
        value: admission-webhook-
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: <randomly-generated-id>
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: minioPvName
        value: minio-pv
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: mysqlPvName
        value: mysql-pv
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: mpi-job/mpi-operator
    name: mpi-operator
  - kustomizeConfig:
      overlays:
      - cognito
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: aws/istio-ingress
    name: istio-ingress
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: istio-system
      - name: origin-header
        value: x-amzn-oidc-data
      - name: custom-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: aws/aws-istio-authz-adaptor
    name: aws-istio-authz-adaptor
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: clusterName
        value: kubeflow-aws
      repoRef:
        name: manifests
        path: aws/aws-alb-ingress-controller
    name: aws-alb-ingress-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: aws/nvidia-device-plugin
    name: nvidia-device-plugin
  plugins:
  - kind: KfAwsPlugin
    metadata:
      name: aws
    spec:
      auth:
        cognito:
          certArn: arn:aws:acm:us-west-2:xxxxx:certificate/xxxxxxxxxxxxx-xxxx
          cognitoAppClientId: xxxxxbxxxxxx
          cognitoUserPoolArn: arn:aws:cognito-idp:us-west-2:xxxxx:userpool/us-west-2_xxxxxx
          cognitoUserPoolDomain: your-user-pool
      region: us-west-2
      roles:
      - eksctl-kubeflow-aws-nodegroup-ng-a2-NodeInstanceRole-xxxxx
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  version: master
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_gcp_iap.yaml", `
# Please set project and email!
apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  # kustomize requires a name.
  name: kfctl-gcp-iap
spec:
  applications:
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/cluster-local-gateway
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "ON"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager        
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: X-Goog-Authenticated-User-Email
      - name: userid-prefix
        value: 'accounts.google.com:'
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: webhookNamePrefix
        value: admission-webhook-
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: X-Goog-Authenticated-User-Email
      - name: userid-prefix
        value: 'accounts.google.com:'
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: injectGcpCredentials
        value: "true"
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: "7439583937720421527"
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      - use-kf-user
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - minioPd
      - application
      parameters:
      - name: minioPd
        value: test1-storage-artifact-store
      - name: minioPvName
        value: minio-pv
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - mysqlPd
      - application
      parameters:
      - name: mysqlPd
        value: test1-storage-metadata-store
      - name: mysqlPvName
        value: mysql-pv
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      - use-kf-user
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - gcp
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      - use-kf-user
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: gcp/cloud-endpoints
    name: cloud-endpoints
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
        # kfctl will set admin to current user account that deploying kubeflow
      - name: admin
        # value: SET_EMAIL
      - name: userid-header
        value: X-Goog-Authenticated-User-Email
      - name: userid-prefix
        value: 'accounts.google.com:'
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: gcp/gpu-driver
    name: gpu-driver
  - kustomizeConfig:
      overlays:
      - managed-cert
      - application
      parameters:
      - name: namespace
        value: istio-system
        # email will be auto-filled.
      - name: ipName
        value: test1-ip
      - name: hostname
        # The value of hostname should be the DNS address for ingress.
        # This will be set automatically during kfctl generate.
        # value: test1.endpoints.SET_PROJECT.cloud.goog
      repoRef:
        name: manifests
        path: gcp/iap-ingress
    name: iap-ingress
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  - kustomizeConfig:
      parameters:
      - name: user
        # kfctl will set user to current user account that deploying kubeflow
        # value: SET_EMAIL
      - name: profile-name
        # kfctl might overwrite profile name
        value: anonymous
      repoRef:
        name: manifests
        path: default-install
    name: default-install
  plugins:
  - kind: KfGcpPlugin
    metadata:
      creationTimestamp: null
      name: gcp
    spec:
      createPipelinePersistentStorage: true
      deploymentManagerConfig:
        repoRef:
          name: manifests
          path: gcp/deployment_manager_configs
      enableWorkloadIdentity: true
      skipInitProject: true
      useBasicAuth: false
      # email should  be set the google account of the person setting up Kubeflow.
      # If its not set kfctl generate will try to set it automatically based on the default
      # gcloud config
      # email: <your_email@gmail.com>
      #
      # Project should be set to the GCP project you want to use.
      # If you run kfctl init --config=<path>/kfctl_gcp_iap.yaml
      # kfctl will try to automatically set it.
      # project: <your project>
      #
      # User can specify which zone to deploy to. If not set, will try to auto-fill
      # this field based on default config in gcloud.
      # zone: us-east1-d
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
    # To get manifest at a PR:
    #uri: https://github.com/kubeflow/manifests/archive/pull/235/head.tar.gz
  version: master
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_gcp_basic_auth.yaml", `
# Please set project and email!
apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-gcp-basic-auth
spec:
  applications:
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/cluster-local-gateway
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "OFF"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager        
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: webhookNamePrefix
        value: admission-webhook-
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: injectGcpCredentials
        value: "true"
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: "2700513155662330975"
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - minioPd
      - application
      parameters:
      - name: minioPd
        value: test1-storage-artifact-store
      - name: minioPvName
        value: minio-pv
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - mysqlPd
      - application
      parameters:
      - name: mysqlPd
        value: test1-storage-metadata-store
      - name: mysqlPvName
        value: mysql-pv
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - gcp
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: ipName
        value: ipName
      - name: hostname
      repoRef:
        name: manifests
        path: gcp/cloud-endpoints
    name: cloud-endpoints
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
      - name: admin
        # email will be auto-filled.
        # value: SET_EMAIL
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: gcp/gpu-driver
    name: gpu-driver
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  - kustomizeConfig:
      parameters:
      - name: ambassadorServiceType
        value: NodePort
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: common/ambassador
    name: ambassador
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: common/basic-auth
    name: basic-auth
  - kustomizeConfig:
      overlays:
      - managed-cert
      - application
      parameters:
      - name: namespace
        value: istio-system
      - name: ipName
        # ipName will be auto-filled based on app name if not set.
        # value: test1-ip
      - name: hostname
        # hostname will be auto-filled if not set.
        # value: <deployName>.endpoints.<project>.cloud.goog
      - name: project
        # Project will be auto-filled.
        # value: SET_PROJECT
      - name: ingressName
        value: envoy-ingress
      - name: issuer
        value: letsencrypt-prod
      repoRef:
        name: manifests
        path: gcp/basic-auth-ingress
    name: basic-auth-ingress
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: default-install
    name: default-install
  plugins:
  - kind: KfGcpPlugin
    metadata:
      creationTimestamp: null
      name: gcp
    spec:
      createPipelinePersistentStorage: true
      deploymentManagerConfig:
        repoRef:
          name: manifests
          path: gcp/deployment_manager_configs
      enableWorkloadIdentity: true
      skipInitProject: true
      useBasicAuth: true
      # email should  be set the google account of the person setting up Kubeflow.
      # If its not set kfctl generate will try to set it automatically based on the default
      # gcloud config
      # email: <your_email@gmail.com>
      #
      # Project should be set to the GCP project you want to use.
      # If you run kfctl init --config=<path>/kfctl_gcp_iap.yaml
      # kfctl will try to automatically set it.
      # project: <your project>
      #
      # User can specify which zone to deploy to. If not set, will try to auto-fill
      # this field based on default config in gcloud.
      # zone: us-east1-d
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  version: master
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_ibm.yaml", `
# This is the config to install Kubeflow on an existing IBM Cloud Kubernetes cluster.
# If the cluster already has istio, comment out the istio install part below.
apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-ibm
spec:
  applications:
  # Istio install. If not needed, comment out istio-crds and istio-install.
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  # This component is the istio resources for Kubeflow (e.g. gateway), not about installing istio.
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/cluster-local-gateway
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "OFF"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: containerRuntimeExecutor
        value: pns
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - ibm-storage-config
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: <randomly-generated-id>
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      - ibm-storage-config
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
      - name: admin
        value: example@kubeflow.org
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  version: master
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_istio_dex.yaml", `
# This is the config to install Kubeflow on an existing K8s cluster, with support
# for multi-user and LDAP auth using Dex.

apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-istio-dex
spec:
  applications:
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  # Istio install. If not needed, comment out istio-crds and istio-install.
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio-1-3-1/istio-crds-1-3-1
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio-1-3-1/istio-install-1-3-1
    name: istio-install
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio-1-3-1/cluster-local-gateway-1-3-1
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "ON"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: istio-system
      - name: userid-header
        value: kubeflow-userid
      - name: oidc_provider
        value: http://dex.auth.svc.cluster.local:5556/dex
      - name: oidc_redirect_uri
        value: /login/oidc
      - name: oidc_auth_url
        value: /dex/auth
      - name: skip_auth_uri
        value: /dex
      - name: client_id
        value: kubeflow-oidc-authservice
      repoRef:
        name: manifests
        path: istio/oidc-authservice
    name: oidc-authservice
  - kustomizeConfig:
      overlays:
      - istio
      parameters:
      - name: namespace
        value: auth
      - name: issuer
        value: http://dex.auth.svc.cluster.local:5556/dex
      - name: client_id
        value: kubeflow-oidc-authservice
      - name: oidc_redirect_uris
        value: '["/login/oidc"]'
      - name: static_email
        value: admin@kubeflow.org
      - name: static_password_hash
        value: $2y$12$ruoM7FqXrpVgaol44eRZW.4HWS8SAvg6KYVVSCIwKQPBmTpCm.EeO
      repoRef:
        name: manifests
        path: dex-auth/dex-crds
    name: dex
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - cert-manager
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: <randomly-generated-id>
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
`)
	th.writeF("/manifests/kfdef/source/master/kfctl_k8s_istio.yaml", `
# This is the config to install Kubeflow on an existing k8s cluster.
# If the cluster already has istio, comment out the istio install part below.

apiVersion: kfdef.apps.kubeflow.org/v1
kind: KfDef
metadata:
  name: kfctl-k8s-istio
spec:
  applications:
  # Istio install. If not needed, comment out istio-crds and istio-install.
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  # This component is the istio resources for Kubeflow (e.g. gateway), not about installing istio.
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/cluster-local-gateway
    name: cluster-local-gateway
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "OFF"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/add-anonymous-user-filter
    name: add-anonymous-user-filter
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-crds
    name: cert-manager-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: kube-system
      repoRef:
        name: manifests
        path: cert-manager/cert-manager-kube-system-resources
    name: cert-manager-kube-system-resources
  - kustomizeConfig:
      overlays:
      - self-signed
      - application
      parameters:
      - name: namespace
        value: cert-manager
      repoRef:
        name: manifests
        path: cert-manager/cert-manager
    name: cert-manager
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kubeflow-roles
    name: kubeflow-roles
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - istio
      - application
      parameters:
      - name: userid-header
        value: kubeflow-userid
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: spark/spark-operator
    name: spark-operator
  - kustomizeConfig:
      overlays:
      - istio
      - application
      - db
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: usageId
        value: <randomly-generated-id>
      - name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-crds
    name: tf-job-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: katib/katib-crds
    name: katib-crds
  - kustomizeConfig:
      overlays:
      - application
      - istio
      repoRef:
        name: manifests
        path: katib/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      overlays:
      - application
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: pipeline/pipeline-visualization-service
    name: pipeline-visualization-service
  - kustomizeConfig:
      overlays:
      - application
      - istio
      parameters:
      - name: admin
        value: johnDoe@acme.com
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  repos:
  - name: manifests
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  version: master
`)
	th.writeK("/manifests/kfdef/source/master", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- kfctl_anthos.yaml
- kfctl_aws.yaml
- kfctl_aws_cognito.yaml
- kfctl_gcp_iap.yaml
- kfctl_gcp_basic_auth.yaml
- kfctl_ibm.yaml
- kfctl_istio_dex.yaml
- kfctl_k8s_istio.yaml
`)
}

func TestKfdefSourceMaster(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kfdef/overlays/master")
	writeKfdefSourceMaster(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kfdef/source/master"
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
